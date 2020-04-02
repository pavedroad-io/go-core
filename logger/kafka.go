package logger

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	stdlog "log"
	"os"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
)

// TopicKey is Fields key to pass topic through WithFields
// Example: log.WithFields(LogFields{TopicKey: "mytopic"}).Infof(...)
const TopicKey string = "topic"

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
	Compression   compressionType
	AckWait       ackWaitType
	ProdFlushFreq time.Duration
	ProdRetryMax  int
	ProdRetryFreq time.Duration
	MetaRetryMax  int
	MetaRetryFreq time.Duration
	EnableTLS     bool
	TLSCfg        *tls.Config
	EnableDebug   bool
}

// KafkaProducer wraps sarama producer with config
type KafkaProducer struct {
	producer sarama.AsyncProducer
	kpConfig ProducerConfiguration
	ceConfig CloudEventsConfiguration
}

// newKafkaProducer returns a kafka producer instance
func newKafkaProducer(
	kpConfig ProducerConfiguration,
	ceConfig CloudEventsConfiguration) (*KafkaProducer, error) {

	if kpConfig.EnableDebug {
		sarama.Logger = stdlog.New(os.Stdout, "[sarama] ", stdlog.LstdFlags)
	}
	cfg := sarama.NewConfig()
	// Consider reading the following two channels with goroutines
	cfg.Producer.Return.Errors = false
	cfg.Producer.Return.Successes = false

	// Override default values if config values are not zero
	if kpConfig.ProdFlushFreq != 0 {
		cfg.Producer.Flush.Frequency = kpConfig.ProdFlushFreq
	}
	if kpConfig.ProdRetryMax != 0 {
		cfg.Producer.Retry.Max = kpConfig.ProdRetryMax
	}
	if kpConfig.ProdRetryFreq != 0 {
		cfg.Producer.Retry.Backoff = kpConfig.ProdRetryFreq
	}
	if kpConfig.MetaRetryMax != 0 {
		cfg.Metadata.Retry.Max = kpConfig.MetaRetryMax
	}
	if kpConfig.MetaRetryFreq != 0 {
		cfg.Metadata.Retry.Backoff = kpConfig.MetaRetryFreq
	}

	switch kpConfig.Partition {
	case HashPartition:
		cfg.Producer.Partitioner = sarama.NewHashPartitioner
	case RoundRobinPartition:
		cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	case RandomPartition:
		fallthrough
	default:
		cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	}

	switch kpConfig.Compression {
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

	switch kpConfig.AckWait {
	case WaitForNone:
		cfg.Producer.RequiredAcks = sarama.NoResponse
	case WaitForAll:
		cfg.Producer.RequiredAcks = sarama.WaitForAll
	case WaitForLocal:
		fallthrough
	default:
		cfg.Producer.RequiredAcks = sarama.WaitForLocal
	}

	if kpConfig.EnableTLS {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = kpConfig.TLSCfg
	}

	producer, err := sarama.NewAsyncProducer(kpConfig.Brokers, cfg)
	if err != nil {
		return &KafkaProducer{}, err
	}

	return &KafkaProducer{
		producer: producer,
		kpConfig: kpConfig,
		ceConfig: ceConfig,
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
	switch kp.kpConfig.Key {
	case FixedKey:
		key = sarama.StringEncoder(kp.kpConfig.KeyName)
	case ExtractedKey:
		if name, ok := msgMap[kp.kpConfig.KeyName].(string); ok {
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

	// capture topic if passed else use default
	topic, ok := msgMap[TopicKey]
	if ok {
		delete(msgMap, TopicKey)
	} else {
		topic = kp.kpConfig.Topic
	}

	// add cloudevents fields like id
	err = kp.ceAddFields(kp.ceConfig, msgMap)
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
		Topic: topic.(string),
		Value: sarama.ByteEncoder(newmsg),
	}
	return nil
}
