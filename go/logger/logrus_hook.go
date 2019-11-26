// from github.com/kenjones-cisco/logrus-kafka-hook/hook.go

package logger

import (
	"errors"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// LogrusHook represents a logrus hook for Kafka
type LogrusHook struct {
	config    ProducerConfiguration
	kp        KafkaProducer
	formatter logrus.Formatter
	levels    []logrus.Level
}

// NewLogrusHook returns a kafka producer hook instance
func NewLogrusHook(config ProducerConfiguration) *LogrusHook {
	return &LogrusHook{
		config:    config,
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
func (h *LogrusHook) WithProducer(producer KafkaProducer) *LogrusHook {
	h.kp = producer

	if h.kp.producer != nil {
		go func() {
			for err := range h.kp.producer.Errors() {
				fmt.Fprintln(os.Stderr, "Producer error", err.Error())
			}
		}()
	}

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
	msg, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	if h.kp.producer == nil {
		return errors.New("no producer defined")
	}

	return h.kp.sendMessage(msg)
}
