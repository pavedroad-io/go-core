// from github.com/amitrai48/logger/zap.go

package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
}

func getEncoder(isJSON bool) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.LevelKey = "subject"
	encoderConfig.MessageKey = "data"
	encoderConfig.CallerKey = ""
	if isJSON {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getZapLevel(level string) zapcore.Level {
	switch level {
	case Info:
		return zapcore.InfoLevel
	case Warn:
		return zapcore.WarnLevel
	case Debug:
		return zapcore.DebugLevel
	case Error:
		return zapcore.ErrorLevel
	case Fatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func newZapLogger(config Configuration) (Logger, error) {
	cores := []zapcore.Core{}

	if config.EnableKafka {
		level := getZapLevel(config.LogLevel)
		// create an async producer
		asyncproducer, err := NewAsyncProducer(config.KafkaProducerCfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "NewAsyncProducer failed", err.Error())
		}
		writer := NewZapWriter(config.KafkaProducerCfg.Topic, asyncproducer)

		var json bool
		if config.KafkaFormat == TypeTextFormat {
			json = false
		} else {
			json = true
		}
		core := zapcore.NewCore(getEncoder(json), writer, level)
		cores = append(cores, core)
	}

	if config.EnableConsole {
		level := getZapLevel(config.LogLevel)
		writer := zapcore.Lock(os.Stdout)
		var json bool
		if config.ConsoleFormat == TypeTextFormat {
			json = false
		} else {
			json = true
		}
		core := zapcore.NewCore(getEncoder(json), writer, level)
		cores = append(cores, core)
	}

	if config.EnableFile {
		level := getZapLevel(config.LogLevel)
		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename: config.FileLocation,
			MaxSize:  100,
			Compress: true,
			MaxAge:   28,
		})
		var json bool
		if config.FileFormat == TypeTextFormat {
			json = false
		} else {
			json = true
		}
		core := zapcore.NewCore(getEncoder(json), writer, level)
		cores = append(cores, core)
	}

	combinedCore := zapcore.NewTee(cores...)

	// AddCallerSkip skips 2 number of callers, this is important else the file that gets
	// logged will always be the wrapped file. In our case zap.go
	logger := zap.New(combinedCore,
		zap.AddCallerSkip(2),
		zap.AddCaller(),
	).Sugar()

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

func (l *zapLogger) Print(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *zapLogger) Printf(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Println(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *zapLogger) Debugf(format string, args ...interface{}) {
	l.sugaredLogger.Debugf(format, args...)
}

func (l *zapLogger) Infof(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Warnf(format string, args ...interface{}) {
	l.sugaredLogger.Warnf(format, args...)
}

func (l *zapLogger) Errorf(format string, args ...interface{}) {
	l.sugaredLogger.Errorf(format, args...)
}

func (l *zapLogger) Fatalf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) Panicf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) WithFields(fields Fields) Logger {
	var f = make([]interface{}, 0)
	for k, v := range fields {
		f = append(f, k)
		f = append(f, v)
	}
	newLogger := l.sugaredLogger.With(f...)
	return &zapLogger{newLogger}
}