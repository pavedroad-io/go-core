// from github.com/amitrai48/logger/logrus.go

package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type logrusLogEntry struct {
	entry *logrus.Entry
}

type logrusLogger struct {
	logger *logrus.Logger
}

// cloudevents formatter
type ceFormatter struct {
	logrus.JSONFormatter
}

// override the Format method for cloudevents
func (f *ceFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	msg, err := f.JSONFormatter.Format(entry)
	if err != nil {
		return nil, err
	}
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	parts := [][]byte{[]byte("{\"id\":\""), []byte(id.String()), []byte("\",")}
	json := bytes.Join(parts, []byte(""))
	return bytes.Replace(msg, []byte("{"), json, 1), nil
}

func getFormatter(isJSON bool, truncate bool) logrus.Formatter {
	if isJSON {
		return &ceFormatter{
			logrus.JSONFormatter{
				TimestampFormat: time.RFC3339,
				FieldMap: logrus.FieldMap{
					logrus.FieldKeyTime:  "time",
					logrus.FieldKeyMsg:   "data",
					logrus.FieldKeyLevel: "subject",
				},
			},
		}
	}
	return &logrus.TextFormatter{
		FullTimestamp:          true,
		DisableLevelTruncation: !truncate,
	}
}

func newLogrusLogger(config Configuration) (Logger, error) {
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, err
	}

	stdOutHandler := os.Stdout
	fileHandler := &lumberjack.Logger{
		Filename: config.FileLocation,
		MaxSize:  100,
		Compress: true,
		MaxAge:   28,
	}
	lLogger := &logrus.Logger{
		Out:       stdOutHandler,
		Formatter: getFormatter(config.ConsoleJSONFormat, false),
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}

	if config.EnableConsole && config.EnableFile {
		lLogger.SetOutput(io.MultiWriter(stdOutHandler, fileHandler))
	} else {
		if config.EnableFile {
			lLogger.SetOutput(fileHandler)
			lLogger.SetFormatter(getFormatter(config.FileJSONFormat, true))
		}
	}

	if config.EnableKafka {
		// create an async producer
		asyncproducer, err := NewAsyncProducer(config.KafkaProducerCfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "NewAsyncProducer failed", err.Error())
		}

		// create the Kafka hook
		hook := NewLogrusHook().WithFormatter(getFormatter(config.KafkaJSONFormat, true)).WithProducer(asyncproducer).WithTopic(config.KafkaProducerCfg.Topic)

		// add the hook
		lLogger.Hooks.Add(hook)
	}

	return &logrusLogger{
		logger: lLogger,
	}, nil
}

func (l *logrusLogger) Print(args ...interface{}) {
	l.logger.Print(args...)
}

func (l *logrusLogger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *logrusLogger) Println(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *logrusLogger) Panicf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *logrusLogger) WithFields(fields Fields) Logger {
	l.logger.SetFormatter(getFormatter(false, true))
	return &logrusLogEntry{
		entry: l.logger.WithFields(convertToLogrusFields(fields)),
	}
}

func (l *logrusLogger) WithCloudEvents() Logger {
	return &logrusLogEntry{
		entry: l.logger.WithFields(convertToLogrusFields(ceFields)),
	}
}

func (l *logrusLogEntry) Print(args ...interface{}) {
	l.entry.Print(args...)
}

func (l *logrusLogEntry) Printf(format string, args ...interface{}) {
	l.entry.Printf(format, args...)
}

func (l *logrusLogEntry) Println(args ...interface{}) {
	l.entry.Println(args...)
}

func (l *logrusLogEntry) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *logrusLogEntry) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logrusLogEntry) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logrusLogEntry) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *logrusLogEntry) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logrusLogEntry) Panicf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logrusLogEntry) WithFields(fields Fields) Logger {
	return &logrusLogEntry{
		entry: l.entry.WithFields(convertToLogrusFields(fields)),
	}
}

func (l *logrusLogEntry) WithCloudEvents() Logger {
	return &logrusLogEntry{
		entry: l.entry.WithFields(convertToLogrusFields(ceFields)),
	}
}

func convertToLogrusFields(fields Fields) logrus.Fields {
	logrusFields := logrus.Fields{}
	for index, val := range fields {
		logrusFields[index] = val
	}
	return logrusFields
}
