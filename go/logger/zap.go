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

func getEncoder(format FormatType) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	switch format {
	case TypeJSONFormat:
		return zapcore.NewJSONEncoder(encoderConfig)
	case TypeTextFormat:
		return zapcore.NewConsoleEncoder(encoderConfig)
	case TypeCEFormat:
		encoderConfig.TimeKey = ceTimeKey
		encoderConfig.LevelKey = ceLevelKey
		encoderConfig.MessageKey = ceMessageKey
		encoderConfig.CallerKey = ""
		return zapcore.NewJSONEncoder(encoderConfig)
	default:
		return nil
	}
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

// Temporary hook for testing
func zapHook(entry zapcore.Entry) error {
	fmt.Printf("ZapHook: entry <%+v>\n", entry)
	return nil
}

func newZapLogger(config Configuration) (Logger, error) {
	cores := []zapcore.Core{}

	if config.EnableKafka {
		level := getZapLevel(config.LogLevel)
		// create an async producer
		kafkaProducer, err := NewKafkaProducer(config.KafkaProducerCfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "NewKafkaProducer failed", err.Error())
		}
		writer := NewZapWriter(config.KafkaProducerCfg, kafkaProducer)
		core := zapcore.NewCore(getEncoder(config.KafkaFormat), writer, level)
		// core = zapcore.RegisterHooks(core, zapHook)
		cores = append(cores, core)
	}

	if config.EnableConsole {
		level := getZapLevel(config.LogLevel)
		writer := zapcore.Lock(os.Stdout)
		core := zapcore.NewCore(getEncoder(config.ConsoleFormat), writer, level)
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

func (l *zapLogger) Print(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *zapLogger) Printf(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Println(args ...interface{}) {
	l.sugaredLogger.Info(fmt.Sprintln(args...))
}

func (l *zapLogger) Debug(args ...interface{}) {
	l.sugaredLogger.Debug(args...)
}

func (l *zapLogger) Debugf(format string, args ...interface{}) {
	l.sugaredLogger.Debugf(format, args...)
}

func (l *zapLogger) Debugln(args ...interface{}) {
	l.sugaredLogger.Debug(fmt.Sprintln(args...))
}

func (l *zapLogger) Info(args ...interface{}) {
	l.sugaredLogger.Info(args...)
}

func (l *zapLogger) Infof(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

func (l *zapLogger) Infoln(args ...interface{}) {
	l.sugaredLogger.Info(fmt.Sprintln(args...))
}

func (l *zapLogger) Warn(args ...interface{}) {
	l.sugaredLogger.Warn(args...)
}

func (l *zapLogger) Warnf(format string, args ...interface{}) {
	l.sugaredLogger.Warnf(format, args...)
}

func (l *zapLogger) Warnln(args ...interface{}) {
	l.sugaredLogger.Warn(fmt.Sprintln(args...))
}

func (l *zapLogger) Error(args ...interface{}) {
	l.sugaredLogger.Error(args...)
}

func (l *zapLogger) Errorf(format string, args ...interface{}) {
	l.sugaredLogger.Errorf(format, args...)
}

func (l *zapLogger) Errorln(args ...interface{}) {
	l.sugaredLogger.Error(fmt.Sprintln(args...))
}

func (l *zapLogger) Fatal(args ...interface{}) {
	l.sugaredLogger.Fatal(args...)
}

func (l *zapLogger) Fatalf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) Fatalln(args ...interface{}) {
	l.sugaredLogger.Fatal(fmt.Sprintln(args...))
}

func (l *zapLogger) Panic(args ...interface{}) {
	l.sugaredLogger.Panic(args...)
}

func (l *zapLogger) Panicf(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

func (l *zapLogger) Panicln(args ...interface{}) {
	l.sugaredLogger.Panic(fmt.Sprintln(args...))
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
