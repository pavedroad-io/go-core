package logger

// Sender provides Kafka producer API with no log features

import "errors"

// SenderConfiguration stores the config for the sender
type SenderConfiguration struct {
	EnableCloudEvents bool
	CloudEventsCfg    CloudEventsConfiguration
	KafkaFormat       FormatType
	KafkaProducerCfg  ProducerConfiguration
	EnableDebug       bool
}

// sender provides object for sender
type sender struct {
	kp *KafkaProducer
	ce *CloudEvents
}

// NewSender returns a Sender instance
func NewSender(config SenderConfiguration) (Sender, error) {
	var errCount int
	var cloudEvents *CloudEvents

	checkProducerConfig(config.KafkaProducerCfg, &errCount)
	if config.EnableCloudEvents {
		checkCETypes(config.CloudEventsCfg, &errCount)
	}
	if errCount > 0 {
		return nil, errors.New("Invalid configuration")
	}
	// create cloud events
	if config.EnableCloudEvents {
		cloudEvents = newCloudEvents(config.CloudEventsCfg)
	}
	// create an async producer
	kafkaProducer, err := newKafkaProducer(config.KafkaProducerCfg,
		cloudEvents, config.CloudEventsCfg)
	if err != nil {
		return nil, err
	}
	// create the sender
	return &sender{
		kp: kafkaProducer,
		ce: cloudEvents,
	}, nil
}

// Sender is the contract for the sender interface
type Sender interface {
	SendCE(value []byte) error

	SendTKV(topic string, key []byte, value []byte) error

	SendMult(topics []string, key []byte, value []byte) error
}

// The following meet the contract for the Sender

// SendCE sends directly to Kafka with Cloud Events
func (s *sender) SendCE(msg []byte) error {
	return s.kp.sendMessage(msg)
}

// SendTKV sends directly to Kafka with no processing
func (s *sender) SendTKV(topic string, key []byte, msg []byte) error {
	return s.kp.sendMessageTKV(topic, key, msg)
}

// SendMult sends message to multiple topics with no processing
func (s *sender) SendMult(topics []string, key []byte, msg []byte) error {
	for _, topic := range topics {
		err := s.SendTKV(topic, key, msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// The following are wrappers with topic added as first argument

func tPrint(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Print(args...)
}

func tPrintf(topic string, format string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Printf(format, args...)
}

func tPrintln(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Println(args...)
}

func tDebug(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Debug(args...)
}

func tDebugf(topic string, format string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Debugf(format, args...)
}

func tDebugln(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Debugln(args...)
}

func tInfo(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Info(args...)
}

func tInfof(topic string, format string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Infof(format, args...)
}

func tInfoln(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Infoln(args...)
}

func tWarn(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Warn(args...)
}

func tWarnf(topic string, format string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Warnf(format, args...)
}

func tWarnln(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Warnln(args...)
}

func tError(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Error(args...)
}

func tErrorf(topic string, format string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Errorf(format, args...)
}

func tErrorln(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Errorln(args...)
}

func tFatal(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Fatal(args...)
}

func tFatalf(topic string, format string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Fatalf(format, args...)
}

func tFatalln(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Fatalln(args...)
}

func tPanic(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Panic(args...)
}

func tPanicf(topic string, format string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Fatalf(format, args...)
}

func tPanicln(topic string, args ...interface{}) {
	topicfield := LogFields{TopicKey: topic}
	logger.WithFields(topicfield).Panicln(args...)
}
