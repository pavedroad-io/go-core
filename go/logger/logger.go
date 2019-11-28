// from github.com/amitrai48/logger/logger.go

package logger

import "errors"

// Fields provides type used in calling WithFields for structured logging
type Fields map[string]interface{}

const (
	//Debug has verbose message
	Debug = "debug"
	//Info is default log level
	Info = "info"
	//Warn is for logging messages about possible issues
	Warn = "warn"
	//Error is for logging errors
	Error = "error"
	//Fatal is for logging fatal messages. The sytem shutsdown after logging the message.
	Fatal = "fatal"
)

// LoggerType provide types of loggers
type LoggerType int8

// Types of loggers
const (
	TypeZapLogger LoggerType = iota
	TypeLogrusLogger
)

// FormatType provides type for logger formats
type FormatType int8

// Types of logger formats
const (
	TypeJSONFormat FormatType = iota
	TypeTextFormat
	TypeCEFormat
)

var (
	errInvalidLoggerType = errors.New("Invalid logger type")
	errNewAsyncProducer  = errors.New("NewAsyncProducer failed")
)

// Logger is our contract for the logger
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

// Configuration stores the config for the logger
type Configuration struct {
	LogLevel             string
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

// NewLogger returns a Logger instance
func NewLogger(config Configuration, loggerType LoggerType) (Logger, error) {
	switch loggerType {
	case TypeZapLogger:
		return newZapLogger(config)
	case TypeLogrusLogger:
		return newLogrusLogger(config)
	default:
		return nil, errInvalidLoggerType
	}
}
