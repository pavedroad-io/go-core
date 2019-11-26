package logger

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
)

type kafkaPartitionType int8

const (
	RandomPartition kafkaPartitionType = iota
	HashPartition
	RoundRobinPartition
)

type kafkaKeyType int8

const (
	LevelKey kafkaKeyType = iota
	TimeSecondKey
	TimeNanoSecondKey
	FixedKey
	ExtractedKey
)

type ProducerConfiguration struct {
	Brokers       []string
	Topic         string
	Partition     kafkaPartitionType
	Key           kafkaKeyType
	KeyName       string
	CloudeventsID ceIDType
	Compression   sarama.CompressionCodec
	Ack           sarama.RequiredAcks
	FlushFreq     time.Duration
	EnableTLS     bool
	TLSCfg        tls.Config
}

// Wrapper around sarama.NewAsyncProducer
type kafkaProducer struct {
	producer sarama.AsyncProducer
	config   ProducerConfiguration
}

func NewKafkaProducer(config ProducerConfiguration) (kafkaProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = config.Ack
	cfg.Producer.Compression = config.Compression
	cfg.Producer.Flush.Frequency = config.FlushFreq * time.Millisecond
	cfg.Producer.Return.Successes = false
	cfg.Producer.Return.Errors = false

	switch config.Partition {
	case HashPartition:
		cfg.Producer.Partitioner = sarama.NewHashPartitioner
	case RandomPartition:
		cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	case RoundRobinPartition:
		cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	default:
		cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	}

	if config.EnableTLS {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = &config.TLSCfg
	}

	producer, err := sarama.NewAsyncProducer(config.Brokers, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "NewAsyncProducer failed", err.Error())
	}

	return kafkaProducer{
		producer: producer,
		config:   config,
	}, err
}

func (kp *kafkaProducer) sendMessage(msg []byte) error {
	var msgMap map[string]interface{}

	// unmarshal message to access fields
	err := json.Unmarshal(msg, &msgMap)
	if err != nil {
		return err
	}
	// can extract data from message fields here
	// set key based on kp config
	var key sarama.Encoder
	switch kp.config.Key {
	case FixedKey:
		key = sarama.StringEncoder(kp.config.KeyName)
	case ExtractedKey:
		// TODO fails if key is not in message
		key = sarama.StringEncoder(msgMap[kp.config.KeyName].(string))
	case TimeSecondKey:
		key = sarama.StringEncoder(strconv.Itoa(int(time.Now().Unix())))
	case TimeNanoSecondKey:
		key = sarama.StringEncoder(strconv.Itoa(int(time.Now().UnixNano())))
	case LevelKey:
		fallthrough
	default:
		key = sarama.StringEncoder(msgMap[ceLevelKey].(string))
	}

	// add cloudevents fields like id
	err = kp.ceAddFields(msgMap)
	if err != nil {
		return err
	}

	// re-marshal message after field manipulation
	newmsg, err := json.Marshal(msgMap)
	if err != nil {
		return err
	}

	kp.producer.Input() <- &sarama.ProducerMessage{
		Key:   key,
		Topic: kp.config.Topic,
		Value: sarama.ByteEncoder(newmsg),
	}
	return nil
}
