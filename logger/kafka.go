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

// TopicKey is LogFields key to pass topic through WithFields
// Example: log.WithFields(LogFields{TopicKey: "mytopic"}).Infof(...)
const TopicKey string = "topic"

// kafkaPartitionType provides kafka partition type
type kafkaPartitionType string

// Types of kafka partitioning to map to sarama
const (
	RandomPartition     kafkaPartitionType = "random" // default
	HashPartition       kafkaPartitionType = "hash"
	RoundRobinPartition kafkaPartitionType = "roundrobin"
)

// kafkaKeyType provides kafka key type
type kafkaKeyType string

// Types of kafka keys to generate
const (
	LevelKey          kafkaKeyType = "level" // default
	TimeSecondKey     kafkaKeyType = "second"
	TimeNanoSecondKey kafkaKeyType = "nanosecond"
	FixedKey          kafkaKeyType = "fixed"
	ExtractedKey      kafkaKeyType = "extracted"
	FunctionKey       kafkaKeyType = "function"
)

// compressionType provides kafka compression type
type compressionType string

// Types of compression to map to sarama
const (
	CompressionNone   compressionType = "none" // default
	CompressionGZIP   compressionType = "gzip"
	CompressionSnappy compressionType = "snappy"
	CompressionLZ4    compressionType = "lz4"
	CompressionZSTD   compressionType = "zstd"
)

// ackWaitType provides kafka acknowledgment wait type
type ackWaitType string

// Types of ack waiting to map to sarama
const (
	// WaitForNone does not wait for any response
	WaitForNone ackWaitType = "none"
	// WaitForLocal waits for only the local commit to succeed
	WaitForLocal ackWaitType = "local" // default
	// WaitForAll waits for all in-sync replicas to commit
	WaitForAll ackWaitType = "all"
)

// FilterFunc func to add/modify/remove message map entries and return kafka key
type FilterFunc func(*map[string]interface{})

// KeyFunc func to return key calculated from kafka message contents
type KeyFunc func(*map[string]interface{}) string

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
	filterFn      FilterFunc
	keyFn         KeyFunc
}

// KafkaProducer wraps sarama producer with config
type KafkaProducer struct {
	producer    sarama.AsyncProducer
	config      ProducerConfiguration
	cloudEvents *CloudEvents
	enableCE    bool
	levelKey    string
}

// newKafkaProducer returns a kafka producer instance
func newKafkaProducer(config ProducerConfiguration, cloudEvents *CloudEvents,
	ceConfig CloudEventsConfiguration) (*KafkaProducer, error) {

	if config.EnableDebug {
		sarama.Logger = stdlog.New(os.Stdout, "[sarama] ", stdlog.LstdFlags)
	}

	cfg := sarama.NewConfig()
	// TODO Consider reading Errors or Successes channels with goroutines
	cfg.Producer.Return.Errors = false
	cfg.Producer.Return.Successes = false

	cfg.Producer.Flush.Frequency = config.ProdFlushFreq
	cfg.Producer.Retry.Max = config.ProdRetryMax
	cfg.Producer.Retry.Backoff = config.ProdRetryFreq
	cfg.Metadata.Retry.Max = config.MetaRetryMax
	cfg.Metadata.Retry.Backoff = config.MetaRetryFreq

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

	var enableCE bool = false
	var levelKey string = "level"
	if cloudEvents != nil {
		enableCE = true
		if ceConfig.SetSubjectLevel {
			levelKey = CESubjectKey
		}
	}

	kp := KafkaProducer{
		// producer:    producer,
		config:      config,
		cloudEvents: cloudEvents,
		enableCE:    enableCE,
		levelKey:    levelKey,
	}

	if len(config.Brokers) == 0 || config.Brokers[0] == "" {
		kp.config.Brokers = defaultProducerConfiguration.Brokers
	}
	if config.Topic == "" {
		kp.config.Topic = defaultProducerConfiguration.Topic
	}
	if config.Key == FixedKey && config.KeyName == "" {
		kp.config.KeyName = defaultProducerConfiguration.KeyName
	}

	producer, err := sarama.NewAsyncProducer(kp.config.Brokers, cfg)
	if err != nil {
		return &KafkaProducer{}, err
	}
	kp.producer = producer

	return &kp, nil
}

// setFilterFn sets the kafka message filter function
func (kp *KafkaProducer) setFilterFn(filterFn FilterFunc) {
	kp.config.filterFn = filterFn
}

// setKeyFn sets the kafka message key function
func (kp *KafkaProducer) setKeyFn(keyFn KeyFunc) {
	kp.config.keyFn = keyFn
}

func (kp *KafkaProducer) getKey(msgMap map[string]interface{},
	key *sarama.Encoder) error {

	// get key based on kp config
	switch kp.config.Key {
	case FixedKey:
		*key = sarama.StringEncoder(kp.config.KeyName)
	case ExtractedKey:
		if name, ok := msgMap[kp.config.KeyName].(string); ok {
			*key = sarama.StringEncoder(name)
			delete(msgMap, kp.config.KeyName)
		} else {
			return errors.New("Extracted key missing")
		}
	case TimeSecondKey:
		*key = sarama.StringEncoder(strconv.Itoa(int(time.Now().Unix())))
	case TimeNanoSecondKey:
		*key = sarama.StringEncoder(strconv.Itoa(int(time.Now().UnixNano())))
	case FunctionKey:
		if kp.config.keyFn != nil {
			*key = sarama.StringEncoder(kp.config.keyFn(&msgMap))
			break
		}
		fallthrough
	case LevelKey:
		fallthrough
	default:
		*key = sarama.StringEncoder(msgMap[kp.levelKey].(string))
	}
	return nil
}

// sendMessage adds key and cloudevents ID before sending message to kafka
func (kp *KafkaProducer) sendMessage(msg []byte) error {
	var msgMap map[string]interface{}

	// unmarshal message to access fields
	err := json.Unmarshal(msg, &msgMap)
	if err != nil {
		return err
	}

	// capture topic if passed else use default
	topic, ok := msgMap[TopicKey]
	if ok {
		delete(msgMap, TopicKey)
	} else {
		topic = kp.config.Topic
	}

	// get kafka key, may delete key from map
	var key sarama.Encoder
	err = kp.getKey(msgMap, &key)
	if err != nil {
		return err
	}

	// filter function performs field manipulation
	if kp.config.filterFn != nil {
		kp.config.filterFn(&msgMap)
	}

	// add cloudevents fields like id (possibly dependent of message)
	// thus must be after all message map manipulation before re-marshal
	if kp.enableCE {
		err = kp.cloudEvents.ceAddFields(msgMap)
		if err != nil {
			return err
		}
	}

	// re-marshal message after field manipulation
	newmsg, err := json.Marshal(msgMap)
	if err != nil {
		return err
	}

	kp.producer.Input() <- &sarama.ProducerMessage{
		Topic: topic.(string),
		Key:   key,
		Value: sarama.ByteEncoder(newmsg),
	}
	return nil
}

// sendMessageTK send msg with topic and key and no processing
func (kp *KafkaProducer) sendMessageTKV(topic string, key []byte,
	value []byte) error {

	kp.producer.Input() <- &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	return nil
}
