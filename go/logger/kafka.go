package logger

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
)

// kafkaPartitionType provides kafka partition type
type kafkaPartitionType string

// Types of kafka partitioning to map to sarama
const (
	RandomPartition     kafkaPartitionType = "random" // default
	HashPartition                          = "hash"
	RoundRobinPartition                    = "roundrobin"
)

// kafkaKeyType provides kafka key type
type kafkaKeyType string

// Types of kafka keys to map to sarama
const (
	LevelKey          kafkaKeyType = "level" // default
	TimeSecondKey                  = "second"
	TimeNanoSecondKey              = "nanosecond"
	FixedKey                       = "fixed"
	ExtractedKey                   = "extracted"
)

// compressionType provides kafka compression type
type compressionType string

// Types of compression to map to sarama
const (
	CompressionNone   compressionType = "none" // default
	CompressionGZIP                   = "gzip"
	CompressionSnappy                 = "snappy"
	CompressionLZ4                    = "lz4"
	CompressionZSTD                   = "zstd"
)

// ackWaitType provides kafka acknowledgment wait type
type ackWaitType string

// Types of ack waiting to map to sarama
const (
	// WaitForNone does not wait for any response
	WaitForNone ackWaitType = "none"
	// WaitForLocal waits for only the local commit to succeed
	WaitForLocal = "local" // default
	// WaitForAll waits for all in-sync replicas to commit
	WaitForAll = "all"
)

// ProducerConfiguration provides kafka producer configuration type
type ProducerConfiguration struct {
	Brokers       []string
	Topic         string
	Partition     kafkaPartitionType
	Key           kafkaKeyType
	KeyName       string
	CloudeventsID ceIDType
	Compression   compressionType
	AckWait       ackWaitType
	FlushFreq     time.Duration
	EnableTLS     bool
	TLSCfg        *tls.Config
}

// KafkaProducer wraps sarama producer with config
type KafkaProducer struct {
	producer sarama.AsyncProducer
	config   ProducerConfiguration
}

// newKafkaProducer returns a kafka producer instance
func newKafkaProducer(config ProducerConfiguration) (*KafkaProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Flush.Frequency = config.FlushFreq * time.Millisecond
	cfg.Producer.Return.Successes = false
	cfg.Producer.Return.Errors = false

	switch config.Partition {
	case HashPartition:
		cfg.Producer.Partitioner = sarama.NewHashPartitioner
	case RoundRobinPartition:
		cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	case RandomPartition:
		fallthrough
	default:
		cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	}

	switch config.Compression {
	case CompressionGZIP:
		cfg.Producer.Compression = sarama.CompressionGZIP
	case CompressionSnappy:
		cfg.Producer.Compression = sarama.CompressionSnappy
	case CompressionLZ4:
		cfg.Producer.Compression = sarama.CompressionLZ4
	case CompressionZSTD:
		cfg.Producer.Compression = sarama.CompressionZSTD
	case CompressionNone:
		fallthrough
	default:
		cfg.Producer.Compression = sarama.CompressionNone
	}

	switch config.AckWait {
	case WaitForNone:
		cfg.Producer.RequiredAcks = sarama.NoResponse
	case WaitForAll:
		cfg.Producer.RequiredAcks = sarama.WaitForAll
	case WaitForLocal:
		fallthrough
	default:
		cfg.Producer.RequiredAcks = sarama.WaitForLocal
	}

	if config.EnableTLS {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = config.TLSCfg
	}

	producer, err := sarama.NewAsyncProducer(config.Brokers, cfg)
	if err != nil {
		return &KafkaProducer{}, err
	}

	return &KafkaProducer{
		producer: producer,
		config:   config,
	}, nil
}

// sendMessage adds key and cloudevents ID before sending message to kafka
func (kp *KafkaProducer) sendMessage(msg []byte) error {
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
		if name, ok := msgMap[kp.config.KeyName].(string); ok {
			key = sarama.StringEncoder(name)
		} else {
			return errors.New("Extracted key missing")
		}
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
