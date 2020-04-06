// Based on github.com/amitrai48/logger/logger.go

package logger

import (
	"errors"
)

// LogFields provided for calls to WithFields for structured logging
type LogFields map[string]interface{}

// PackageType provided to select underlying log package
type PackageType string

// Supported log packages
const (
	ZapType    PackageType = "zap"
	LogrusType PackageType = "logrus"
)

// LevelType provided to select log level
type LevelType string

// Supported log levels
const (
	DebugType LevelType = "debug"
	InfoType  LevelType = "info" // default
	WarnType  LevelType = "warn"
	ErrorType LevelType = "error"
	FatalType LevelType = "fatal"
	PanicType LevelType = "panic"
)

// FormatType provided to select logger format
type FormatType string

// Types of logger formats
const (
	JSONFormat FormatType = "json"
	TextFormat FormatType = "text" // default
	CEFormat   FormatType = "cloudevents"
)

// ConsoleType provided to select logger format
type ConsoleType string

// Types of logger formats
const (
	Stdout ConsoleType = "stdout" // default
	Stderr ConsoleType = "stderr"
)

// Configuration stores the config for the logger
type Configuration struct {
	LogPackage        PackageType
	LogLevel          LevelType
	EnableTimeStamps  bool
	EnableColorLevels bool
	EnableCloudEvents bool
	CloudEventsCfg    CloudEventsConfiguration
	EnableKafka       bool
	KafkaFormat       FormatType
	KafkaProducerCfg  ProducerConfiguration
	EnableConsole     bool
	ConsoleFormat     FormatType
	ConsoleWriter     ConsoleType
	EnableFile        bool
	FileFormat        FormatType
	FileLocation      string
	EnableRotation    bool
	RotationCfg       RotationConfiguration
	EnableDebug       bool
}

// NewLogger returns a Logger instance
func NewLogger(config Configuration) (Logger, error) {
	switch config.LogPackage {
	case ZapType:
		return newZapLogger(config)
	case LogrusType:
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

	WithFields(keyValues LogFields) Logger

	WithKafkaFilterFn(filter FilterFunc) Logger

	WithKafkaKeyFn(filter KeyFunc) Logger
}
