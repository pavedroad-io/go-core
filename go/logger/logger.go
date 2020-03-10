// Based on github.com/amitrai48/logger/logger.go

package logger

import "errors"

// Fields provided for calls to WithFields for structured logging
type Fields map[string]interface{}

// LevelType provided to select log level
type LevelType string

// Supported log levels
const (
	DebugType LevelType = "debug"
	InfoType            = "info" // default
	WarnType            = "warn"
	ErrorType           = "error"
	FatalType           = "fatal"
	PanicType           = "panic"
)

// FormatType provided to select logger format
type FormatType int8

// Types of logger formats
const (
	JSONFormat FormatType = iota
	TextFormat            // default
	CEFormat              // cloudevents
)

// Configuration stores the config for the logger
type Configuration struct {
	LogLevel             LevelType
	EnableCloudEvents    bool
	EnableKafka          bool
	KafkaFormat          FormatType
	KafkaProducerCfg     ProducerConfiguration
	EnableConsole        bool
	ConsoleFormat        FormatType
	ConsoleLevelTruncate bool
	EnableFile           bool
	FileFormat           FormatType
	FileLevelTruncate    bool
	FileLocation         string
}

// LogType provided to select underlying log package
type LogType int8

// Supported log packages
const (
	Zap LogType = iota
	Logrus
)

// NewLogger returns a Logger instance
func NewLogger(config Configuration, logType LogType) (Logger, error) {
	switch logType {
	case Zap:
		return newZapLogger(config)
	case Logrus:
		return newLogrusLogger(config)
	default:
		return nil, errors.New("Invalid log type")
	}
}

// Logger is the contract for the logger interface
type Logger interface {
	Print(args ...interface{})

	Printf(format string, args ...interface{})

	Println(args ...interface{})

	Debug(args ...interface{})

	Debugf(format string, args ...interface{})

	Debugln(args ...interface{})

	Info(args ...interface{})

	Infof(format string, args ...interface{})

	Infoln(args ...interface{})

	Warn(args ...interface{})

	Warnf(format string, args ...interface{})

	Warnln(args ...interface{})

	Error(args ...interface{})

	Errorf(format string, args ...interface{})

	Errorln(args ...interface{})

	Fatal(args ...interface{})

	Fatalf(format string, args ...interface{})

	Fatalln(args ...interface{})

	Panic(args ...interface{})

	Panicf(format string, args ...interface{})

	Panicln(args ...interface{})

	WithFields(keyValues Fields) Logger
}
