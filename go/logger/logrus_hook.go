// from github.com/kenjones-cisco/logrus-kafka-hook/hook.go

package logger

import (
	"errors"
	"fmt"
	"os"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

// Object represents a logrus hook for Kafka
type LogrusHook struct {
	formatter logrus.Formatter
	levels    []logrus.Level
	producer  sarama.AsyncProducer
	topic     string
}

// New creates a new logrus hook for Kafka
//
// Defaults:
//
//		Formatter: *logrus.TextFormatter*
//		Levels: *logrus.AllLevels*
//

func NewLogrusHook() *LogrusHook {
	return &LogrusHook{
		formatter: new(logrus.TextFormatter),
		levels:    logrus.AllLevels,
	}
}

// WithFormatter adds a formatter to the created Hook
func (h *LogrusHook) WithFormatter(formatter logrus.Formatter) *LogrusHook {
	h.formatter = formatter
	return h
}

// WithLevels adds levels to the created Hook
func (h *LogrusHook) WithLevels(levels []logrus.Level) *LogrusHook {
	h.levels = levels
	return h
}

// WithProducer adds a producer to the created Hook
func (h *LogrusHook) WithProducer(producer sarama.AsyncProducer) *LogrusHook {
	h.producer = producer

	if producer != nil {
		go func() {
			for err := range producer.Errors() {
				fmt.Fprintln(os.Stderr, "[ERROR]", err)
			}
		}()
	}

	return h
}

// WithTopic adds a topic to the created Hook
func (h *LogrusHook) WithTopic(topic string) *LogrusHook {
	h.topic = topic
	return h
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
	var key sarama.Encoder

	if t, err := entry.Time.MarshalBinary(); err == nil {
		key = sarama.ByteEncoder(t)
	} else {
		key = sarama.StringEncoder(entry.Level.String())
	}

	msg, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	if h.producer == nil {
		return errors.New("no producer defined")
	}

	h.producer.Input() <- &sarama.ProducerMessage{
		Topic: h.topic,
		Key:   key,
		Value: sarama.ByteEncoder(msg),
	}

	return nil
}
