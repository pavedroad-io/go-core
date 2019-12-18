// Based on github.com/amitrai48/logger/zap.go

package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// zapLogger represents a zap sugar logger
type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
}

// getEncoder returns a zap encoder
func getEncoder(format FormatType) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	switch format {
	case JSONFormat:
		return zapcore.NewJSONEncoder(encoderConfig)
	case CEFormat:
		encoderConfig.TimeKey = ceTimeKey
		encoderConfig.LevelKey = ceLevelKey
		encoderConfig.MessageKey = ceMessageKey
		encoderConfig.CallerKey = ""
		return zapcore.NewJSONEncoder(encoderConfig)
	case TextFormat:
		fallthrough
	default:
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
}

// getZapLevel converts log level to zap log level
func getZapLevel(level LevelType) zapcore.Level {
	switch level {
	case Debug:
		return zapcore.DebugLevel
	case Info:
		return zapcore.InfoLevel
	case Warn:
		return zapcore.WarnLevel
	case Error:
		return zapcore.ErrorLevel
	case Fatal:
		return zapcore.FatalLevel
	case Panic:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// zapHook is a temporary hook for testing
func zapHook(entry zapcore.Entry) error {
	fmt.Printf("ZapHook: entry <%+v>\n", entry)
	return nil
}

// newZapLogger returns a zap logger instance
func newZapLogger(config Configuration) (Logger, error) {
	level := getZapLevel(config.LogLevel)
	cores := []zapcore.Core{}

	if config.EnableKafka {
		writer, err := newZapWriter(config.KafkaProducerCfg)
		if err != nil {
			return nil, err
		}
		core := zapcore.NewCore(getEncoder(config.KafkaFormat), writer, level)
		// core = zapcore.RegisterHooks(core, zapHook)
		cores = append(cores, core)
	}

	if config.EnableConsole {
		writer := zapcore.Lock(os.Stdout)
		core := zapcore.NewCore(getEncoder(config.ConsoleFormat), writer, level)
		cores = append(cores, core)
	}

	if config.EnableFile {
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename: config.FileLocation,
			MaxSize:  100,
			Compress: true,
			MaxAge:   28,
		})
		core := zapcore.NewCore(getEncoder(config.FileFormat), writer, level)
		cores = append(cores, core)
	}

	combinedCore := zapcore.NewTee(cores...)
	logger := zap.New(combinedCore).Sugar()

	if config.EnableCloudEvents {
		zaplogger := &zapLogger{
			sugaredLogger: logger,
		}
		return zaplogger.WithFields(ceFields), nil
	}
	return &zapLogger{
		sugaredLogger: logger,
	}, nil
}

// The following methods meet the contract for the logger interface

func (l *zapLogger) Print(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *zapLogger) Printf(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Println(args ...interface{}) {
	l.sugaredLogger.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func (l *zapLogger) Debug(args ...interface{}) {
	l.sugaredLogger.Debug(args...)
}

func (l *zapLogger) Debugf(format string, args ...interface{}) {
	l.sugaredLogger.Debugf(format, args...)
}

func (l *zapLogger) Debugln(args ...interface{}) {
	l.sugaredLogger.Debug(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func (l *zapLogger) Info(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *zapLogger) Infof(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Infoln(args ...interface{}) {
	l.sugaredLogger.Info(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func (l *zapLogger) Warn(args ...interface{}) {
	l.sugaredLogger.Warn(args...)
}

func (l *zapLogger) Warnf(format string, args ...interface{}) {
	l.sugaredLogger.Warnf(format, args...)
}

func (l *zapLogger) Warnln(args ...interface{}) {
	l.sugaredLogger.Warn(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func (l *zapLogger) Error(args ...interface{}) {
	l.sugaredLogger.Error(args...)
}

func (l *zapLogger) Errorf(format string, args ...interface{}) {
	l.sugaredLogger.Errorf(format, args...)
}

func (l *zapLogger) Errorln(args ...interface{}) {
	l.sugaredLogger.Error(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func (l *zapLogger) Fatal(args ...interface{}) {
	l.sugaredLogger.Fatal(args...)
}

func (l *zapLogger) Fatalf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) Fatalln(args ...interface{}) {
	l.sugaredLogger.Fatal(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

func (l *zapLogger) Panic(args ...interface{}) {
	l.sugaredLogger.Panic(args...)
}

func (l *zapLogger) Panicf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) Panicln(args ...interface{}) {
	l.sugaredLogger.Panic(strings.TrimRight(fmt.Sprintln(args...), "\n"))
}

// WithFields add fixed fileds to each log record
func (l *zapLogger) WithFields(fields Fields) Logger {
	var f = make([]interface{}, 0)
	for k, v := range fields {
		f = append(f, k)
		f = append(f, v)
	}
	newLogger := l.sugaredLogger.With(f...)
	return &zapLogger{newLogger}
}
