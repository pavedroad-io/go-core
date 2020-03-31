// Based on github.com/kenjones-cisco/logrus-kafka-hook/hook.go

package logger

import (
	"errors"

	"github.com/sirupsen/logrus"
)

// LogrusHook provides a logrus hook for Kafka
type LogrusHook struct {
	kp        *KafkaProducer
	formatter logrus.Formatter
	levels    []logrus.Level
}

// newLogrusHook returns a kafka producer hook instance
func newLogrusHook(
	kpcfg ProducerConfiguration,
	cecfg CloudEventsConfiguration,
	fmt logrus.Formatter) (*LogrusHook, error) {

	// create an async producer
	kafkaProducer, err := newKafkaProducer(kpcfg, cecfg)
	if err != nil {
		return nil, err
	}

	// create the Kafka hook
	return &LogrusHook{
		kp:        kafkaProducer,
		formatter: fmt,
		levels:    logrus.AllLevels,
	}, nil
}

/*
The following two methods meet the logrus Hook interface contract:
type Hook interface {
	Levels() []Level
	Fire(*Entry) error
}
*/

// Levels returns all log levels that are enabled for writing messages to Kafka
func (h *LogrusHook) Levels() []logrus.Level {
	return h.levels
}

// Fire writes the entry as a message on Kafka
func (h *LogrusHook) Fire(entry *logrus.Entry) error {
	msg, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	if h.kp.producer == nil {
		return errors.New("No producer defined")
	}

	return h.kp.sendMessage(msg)
}
