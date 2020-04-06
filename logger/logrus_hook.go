// Based on github.com/kenjones-cisco/logrus-kafka-hook/hook.go

package logger

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// LogrusKafkaHook provides a kafka producer hook
type LogrusKafkaHook struct {
	kp        *KafkaProducer
	formatter logrus.Formatter
	levels    []logrus.Level
}

// newLogrusKafkaHook returns a kafka producer hook instance
func newLogrusKafkaHook(
	kpcfg ProducerConfiguration,
	cecfg CloudEventsConfiguration,
	fmt logrus.Formatter) (*LogrusKafkaHook, error) {

	// create an async producer
	kafkaProducer, err := newKafkaProducer(kpcfg, cecfg)
	if err != nil {
		return nil, err
	}

	// create the Kafka hook
	return &LogrusKafkaHook{
		kp:        kafkaProducer,
		formatter: fmt,
		levels:    logrus.AllLevels,
	}, nil
}

// Levels returns all log levels that are enabled
func (h *LogrusKafkaHook) Levels() []logrus.Level {
	return h.levels
}

// Fire writes the entry as a message on Kafka
func (h *LogrusKafkaHook) Fire(entry *logrus.Entry) error {
	msg, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}

	if h.kp.producer == nil {
		return errors.New("No producer defined")
	}

	return h.kp.sendMessage(msg)
}

// LogrusConsoleHook provides a console hook
type LogrusConsoleHook struct {
	out       io.Writer
	formatter logrus.Formatter
	levels    []logrus.Level
}

// newLogrusConsoleHook returns a debug hook instance
func newLogrusConsoleHook(out io.Writer,
	fmt logrus.Formatter) *LogrusConsoleHook {
	// return the console hook
	return &LogrusConsoleHook{
		out:       out,
		formatter: fmt,
		levels:    logrus.AllLevels,
	}
}

// Levels returns all log levels that are enabled
func (h *LogrusConsoleHook) Levels() []logrus.Level {
	return h.levels
}

// Fire writes the log message exactly the same as logrus console
func (h *LogrusConsoleHook) Fire(entry *logrus.Entry) error {
	msg, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	fmt.Fprint(h.out, string(msg))
	return nil
}

// LogrusDebugHook provides a debug hook
type LogrusDebugHook struct {
	levels []logrus.Level
}

// newLogrusDebugHook returns a debug hook instance
func newLogrusDebugHook() *LogrusDebugHook {
	// return the debug hook
	return &LogrusDebugHook{
		levels: logrus.AllLevels,
	}
}

// Levels returns all log levels that are enabled
func (h *LogrusDebugHook) Levels() []logrus.Level {
	return h.levels
}

// Fire prints the entry
func (h *LogrusDebugHook) Fire(entry *logrus.Entry) error {
	fmt.Fprintf(os.Stderr, "entry <%+v>\n", entry)
	return nil
}
