// Based on github.com/amitrai48/logger/logrus.go

package logger

import (
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// logrusLogger provides object for logrus logger
type logrusLogger struct {
	logger *logrus.Logger
}

// logrusLogEntry provides object for logrus logger with Entry set by WithFields
type logrusLogEntry struct {
	entry *logrus.Entry
}

// ceFormatter provides the cloudevents formatter type
type ceFormatter struct {
	logrus.JSONFormatter
}

// Format overrides the Format method for cloudevents
func (f *ceFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// TODO may no longer need this as msg now modified in sendMessage
	msg, err := f.JSONFormatter.Format(entry)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// getFormatter returns a logrus formatter
func getFormatter(format FormatType, truncate bool) logrus.Formatter {
	switch format {
	case JSONFormat:
		return &logrus.JSONFormatter{}
	case CEFormat:
		return &ceFormatter{
			logrus.JSONFormatter{
				TimestampFormat: time.RFC3339,
				FieldMap: logrus.FieldMap{
					logrus.FieldKeyTime:  ceTimeKey,
					logrus.FieldKeyMsg:   ceMessageKey,
					logrus.FieldKeyLevel: ceLevelKey,
				},
			},
		}
	case TextFormat:
		fallthrough
	default:
		return &logrus.TextFormatter{
			FullTimestamp:          true,
			DisableLevelTruncation: !truncate,
		}
	}
}

// newLogrusLogger return a logrus logger instance
func newLogrusLogger(config Configuration) (Logger, error) {
	level, err := logrus.ParseLevel(string(config.LogLevel))
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
		Formatter: getFormatter(config.ConsoleFormat, config.ConsoleLevelTruncate),
		Hooks:     make(logrus.LevelHooks),
		Level:     level,
	}

	if config.EnableConsole && config.EnableFile {
		lLogger.SetOutput(io.MultiWriter(stdOutHandler, fileHandler))
	} else {
		if config.EnableFile {
			lLogger.SetOutput(fileHandler)
			lLogger.SetFormatter(getFormatter(config.FileFormat, config.FileLevelTruncate))
		}
	}

	if config.EnableKafka {
		hook, err := newLogrusHook(config.KafkaProducerCfg, getFormatter(config.KafkaFormat, true))
		if err != nil {
			return nil, err
		}

		// add the hook
		lLogger.Hooks.Add(hook)
	}

	if config.EnableCloudEvents {
		logruslogger := &logrusLogger{
			logger: lLogger,
		}
		return logruslogger.WithFields(ceFields), nil
	}

	return &logrusLogger{
		logger: lLogger,
	}, nil
}

// The following meet the contract for the logger

func (l *logrusLogger) Print(args ...interface{}) {
	l.logger.Print(args...)
}

func (l *logrusLogger) Printf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}

func (l *logrusLogger) Println(args ...interface{}) {
	l.logger.Println(args...)
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *logrusLogger) Debugln(args ...interface{}) {
	l.logger.Debugln(args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *logrusLogger) Infoln(args ...interface{}) {
	l.logger.Infoln(args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *logrusLogger) Warnln(args ...interface{}) {
	l.logger.Warnln(args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *logrusLogger) Errorln(args ...interface{}) {
	l.logger.Errorln(args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *logrusLogger) Fatalln(args ...interface{}) {
	l.logger.Fatalln(args...)
}

func (l *logrusLogger) Panic(args ...interface{}) {
	l.logger.Panic(args...)
}

func (l *logrusLogger) Panicf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *logrusLogger) Panicln(args ...interface{}) {
	l.logger.Panicln(args...)
}

// WithFields converts logger to logger with Entry
func (l *logrusLogger) WithFields(fields Fields) Logger {
	return &logrusLogEntry{
		entry: l.logger.WithFields(convertToLogrusFields(fields)),
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

func (l *logrusLogEntry) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

func (l *logrusLogEntry) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *logrusLogEntry) Debugln(args ...interface{}) {
	l.entry.Debugln(args...)
}

func (l *logrusLogEntry) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l *logrusLogEntry) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logrusLogEntry) Infoln(args ...interface{}) {
	l.entry.Infoln(args...)
}

func (l *logrusLogEntry) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l *logrusLogEntry) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logrusLogEntry) Warnln(args ...interface{}) {
	l.entry.Warnln(args...)
}

func (l *logrusLogEntry) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l *logrusLogEntry) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *logrusLogEntry) Errorln(args ...interface{}) {
	l.entry.Errorln(args...)
}

func (l *logrusLogEntry) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

func (l *logrusLogEntry) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logrusLogEntry) Fatalln(args ...interface{}) {
	l.entry.Fatalln(args...)
}

func (l *logrusLogEntry) Panic(args ...interface{}) {
	l.entry.Panic(args...)
}

func (l *logrusLogEntry) Panicf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logrusLogEntry) Panicln(args ...interface{}) {
	l.entry.Panicln(args...)
}

// WithFields add more fields to logger with Entry
func (l *logrusLogEntry) WithFields(fields Fields) Logger {
	return &logrusLogEntry{
		entry: l.entry.WithFields(convertToLogrusFields(fields)),
	}
}

// convertToLogrusFields converts fields to logrus type
func convertToLogrusFields(fields Fields) logrus.Fields {
	logrusFields := logrus.Fields{}
	for index, val := range fields {
		logrusFields[index] = val
	}
	return logrusFields
}
