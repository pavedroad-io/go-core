// Inspired by github.com/amitrai48/logger/logrus.go

package logger

import (
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// logrusLogger provides object for logrus logger
type logrusLogger struct {
	logger    *logrus.Logger
	kafkaHook *LogrusKafkaHook
}

// logrusLogEntry provides object for logrus logger with Entry set by WithFields
type logrusLogEntry struct {
	entry     *logrus.Entry
	kafkaHook *LogrusKafkaHook
}

// ceFormatter provides wrapper for the JSONFormatter (to insert CE fields)
type ceFormatter struct {
	logrus.JSONFormatter
	fields logrus.Fields
}

// Format meets the interface for the logrus formatter
func (ce *ceFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// CE fields are added here, not by using WithFields at logger level
	// make a deep copy of entry with the CE fields to format
	// modifying entry directly would affect other formatters
	ceEntry := entry.WithFields(ce.fields)
	ceEntry.Level = entry.Level
	ceEntry.Message = entry.Message
	msg, err := ce.JSONFormatter.Format(ceEntry)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// getFormatter returns a logrus formatter
func getFormatter(format FormatType, config LoggerConfiguration,
	fields LogFields) logrus.Formatter {

	switch format {
	case JSONFormat:
		return &logrus.JSONFormatter{
			DisableTimestamp: !config.EnableTimeStamps,
			TimestampFormat:  time.RFC3339,
		}
	case CEFormat:
		// Change keys for cloudevents
		fieldmap := logrus.FieldMap{}
		if config.EnableCloudEvents {
			fieldmap[logrus.FieldKeyMsg] = CEDataKey
			if config.CloudEventsCfg.SetSubjectLevel {
				fieldmap[logrus.FieldKeyLevel] = CESubjectKey
			}
		}
		ceFields := logrus.Fields{}
		for key, val := range fields {
			ceFields[key] = val
		}
		return &ceFormatter{
			logrus.JSONFormatter{
				DisableTimestamp: !config.EnableTimeStamps,
				TimestampFormat:  time.RFC3339,
				FieldMap:         fieldmap,
			},
			ceFields,
		}
	case TextFormat:
		fallthrough
	default:
		formatter := logrus.TextFormatter{
			DisableTimestamp: !config.EnableTimeStamps,
			TimestampFormat:  time.RFC3339,
			FullTimestamp:    true,
		}
		// these settings create identical output for ttys and logs
		if config.EnableColorLevels {
			formatter.ForceColors = true
		} else {
			formatter.DisableColors = true
		}
		return &formatter
	}
}

// newLogrusLogger return a logrus logger instance
func newLogrusLogger(config LoggerConfiguration) (Logger, error) {
	var kafkaHook *LogrusKafkaHook
	var cloudEvents *CloudEvents
	var fields LogFields

	logLevel := config.LogLevel
	if logLevel == "" {
		logLevel = defaultLoggerConfiguration.LogLevel
	}
	level, err := logrus.ParseLevel(string(logLevel))
	if err != nil {
		return nil, err
	}

	// set default to discard for kafka only, otherwise overridden
	lLogger := &logrus.Logger{
		Out:          ioutil.Discard,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        level,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}

	if config.EnableCloudEvents {
		cloudEvents = newCloudEvents(config.CloudEventsCfg)
		fields = cloudEvents.fields
	}

	if config.EnableFile {
		var fwriter io.Writer
		var err error
		fileLocation := config.FileLocation
		if fileLocation == "" {
			fileLocation = defaultLoggerConfiguration.FileLocation
		}
		if config.EnableRotation {
			fwriter = rotationLogger(fileLocation, config.RotationCfg)
		} else {
			fwriter, err = os.OpenFile(fileLocation,
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
		}
		lLogger.SetOutput(fwriter)
		lLogger.SetFormatter(getFormatter(config.FileFormat, config, fields))
	} else if config.EnableConsole {
		var cwriter io.Writer
		if debugCapture != nil {
			cwriter = debugCapture
		} else if config.ConsoleWriter == Stderr {
			cwriter = os.Stderr
		} else {
			cwriter = os.Stdout
		}
		formatter := getFormatter(config.ConsoleFormat, config, fields)
		if config.EnableFile {
			// use hook to provide separate formatting for console
			hook := newLogrusConsoleHook(cwriter, formatter)
			lLogger.Hooks.Add(hook)
		} else {
			// otherwise override logger defaults with console settings
			lLogger.SetOutput(cwriter)
			lLogger.SetFormatter(formatter)
		}
	}

	if config.EnableKafka {
		formatter := getFormatter(config.KafkaFormat, config, fields)
		kafkaHook, err = newLogrusKafkaHook(config.KafkaProducerCfg,
			cloudEvents, config.CloudEventsCfg, formatter)
		if err != nil {
			return nil, err
		}
		// add the hook
		lLogger.Hooks.Add(kafkaHook)
	}

	if config.EnableDebug {
		// use hook to provide log entry printing
		hook := &LogrusDebugHook{}
		lLogger.Hooks.Add(hook)
	}

	return &logrusLogger{
		logger:    lLogger,
		kafkaHook: kafkaHook,
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

// WithFields adds more fields to logger, uses logrusLogEntry
func (l *logrusLogger) WithFields(fields LogFields) Logger {
	return &logrusLogEntry{
		entry: l.logger.WithFields(convertToLogrusFields(fields)),
	}
}

// WithKafkaFilterFn adds a filter function for each kafka record
func (l *logrusLogger) WithKafkaFilterFn(filterFn FilterFunc) Logger {
	l.kafkaHook.kp.config.filterFn = filterFn
	return l
}

// WithKafkaKeyFn adds a key function for each kafka record
func (l *logrusLogger) WithKafkaKeyFn(keyFn KeyFunc) Logger {
	l.kafkaHook.kp.config.keyFn = keyFn
	return l
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

// WithFields adds more fields to logger with Entry
func (l *logrusLogEntry) WithFields(fields LogFields) Logger {
	return &logrusLogEntry{
		entry: l.entry.WithFields(convertToLogrusFields(fields)),
	}
}

// WithKafkaFilterFn adds a filter function for each kafka record
func (l *logrusLogEntry) WithKafkaFilterFn(filterFn FilterFunc) Logger {
	l.kafkaHook.kp.config.filterFn = filterFn
	return l
}

// WithKafkaKeyFn adds a key function for each kafka record
func (l *logrusLogEntry) WithKafkaKeyFn(keyFn KeyFunc) Logger {
	l.kafkaHook.kp.config.keyFn = keyFn
	return l
}

// convertToLogrusFields converts fields to logrus type
func convertToLogrusFields(fields LogFields) logrus.Fields {
	logrusFields := logrus.Fields{}
	for index, val := range fields {
		logrusFields[index] = val
	}
	return logrusFields
}
