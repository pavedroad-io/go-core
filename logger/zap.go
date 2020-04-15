// Inspired by github.com/amitrai48/logger/zap.go

package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// zapLogger represents a zap sugar logger
type zapLogger struct {
	sugaredLogger *zap.SugaredLogger
	kafkaWriter   *ZapKafkaWriter
}

// ceEncoder provides wrapper for the JSONEncoder (to insert CE fields)
type ceEncoder struct {
	zapcore.Encoder
	fields []zapcore.Field
}

// Clone meets the interface for the zapcore encoder
func (ce *ceEncoder) Clone() zapcore.Encoder {
	return &ceEncoder{
		ce.Encoder.Clone(),
		ce.fields,
	}
}

// EncodeEntry meets the interface for the zapcore encoder
func (ce *ceEncoder) EncodeEntry(entry zapcore.Entry,
	fields []zapcore.Field) (*buffer.Buffer, error) {
	// CE fields are added here, not by using WithFields
	fields = append(fields, ce.fields...)
	return ce.Encoder.EncodeEntry(entry, fields)
}

// getEncoder returns a zap encoder
func getEncoder(format FormatType, config LoggerConfiguration,
	fields LogFields) zapcore.Encoder {

	encoderConfig := zap.NewProductionEncoderConfig()
	if config.EnableTimeStamps {
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		encoderConfig.TimeKey = CETimeKey
	} else {
		encoderConfig.TimeKey = zapcore.OmitKey
	}
	encoderConfig.NameKey = zapcore.OmitKey
	encoderConfig.CallerKey = zapcore.OmitKey
	encoderConfig.StacktraceKey = zapcore.OmitKey

	switch format {
	case JSONFormat:
		return zapcore.NewJSONEncoder(encoderConfig)
	case CEFormat:
		// Change keys for cloudevents
		if config.EnableCloudEvents {
			encoderConfig.MessageKey = CEDataKey
			if config.CloudEventsCfg.SetSubjectLevel {
				encoderConfig.LevelKey = CESubjectKey
			}
		}
		ceFields := []zapcore.Field{}
		for key, val := range fields {
			ceField := zapcore.Field{
				Key:    key,
				Type:   zapcore.StringType,
				String: val.(string),
			}
			ceFields = append(ceFields, ceField)
		}
		return &ceEncoder{
			zapcore.NewJSONEncoder(encoderConfig),
			ceFields,
		}
	case TextFormat:
		fallthrough
	default:
		if config.EnableColorLevels {
			encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
}

// getZapLevel converts log level to zap log level
func getZapLevel(level LevelType) zapcore.Level {
	switch level {
	case DebugType:
		return zapcore.DebugLevel
	case InfoType:
		return zapcore.InfoLevel
	case WarnType:
		return zapcore.WarnLevel
	case ErrorType:
		return zapcore.ErrorLevel
	case FatalType:
		return zapcore.FatalLevel
	case PanicType:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// zapDebugHook is a hook for testing
func zapDebugHook(entry zapcore.Entry) error {
	fmt.Fprintf(os.Stderr, "%+v\n", entry)
	return nil
}

// newZapLogger returns a zap logger instance
func newZapLogger(config LoggerConfiguration) (Logger, error) {
	var kafkaWriter *ZapKafkaWriter
	var cloudEvents *CloudEvents
	var fields LogFields
	var err error
	level := getZapLevel(config.LogLevel)
	cores := []zapcore.Core{}

	if config.EnableCloudEvents {
		cloudEvents = newCloudEvents(config.CloudEventsCfg)
		fields = cloudEvents.fields
	}

	if config.EnableDebug {
		writer := zapcore.Lock(zapcore.AddSync(ioutil.Discard))
		encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)
		core = zapcore.RegisterHooks(core, zapDebugHook)
		cores = append(cores, core)
	}

	if config.EnableKafka {
		kafkaWriter, err = newZapKafkaWriter(config.KafkaProducerCfg,
			cloudEvents, config.CloudEventsCfg)
		if err != nil {
			return nil, err
		}
		encoder := getEncoder(config.KafkaFormat, config, fields)
		core := zapcore.NewCore(encoder, kafkaWriter, level)
		cores = append(cores, core)
	}

	if config.EnableConsole {
		var cwriter io.Writer
		if config.ConsoleWriter == Stderr {
			cwriter = os.Stderr
		} else {
			cwriter = os.Stdout
		}
		writer := zapcore.Lock(zapcore.AddSync(cwriter))
		encoder := getEncoder(config.ConsoleFormat, config, fields)
		core := zapcore.NewCore(encoder, writer, level)
		cores = append(cores, core)
	}

	if config.EnableFile {
		var fwriter io.Writer
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
		writer := zapcore.AddSync(fwriter)
		encoder := getEncoder(config.FileFormat, config, fields)
		core := zapcore.NewCore(encoder, writer, level)
		cores = append(cores, core)
	}

	combinedCore := zapcore.NewTee(cores...)
	logger := zap.New(combinedCore).Sugar()
	defer logger.Sync()

	return &zapLogger{
		sugaredLogger: logger,
		kafkaWriter:   kafkaWriter,
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

// WithFields adds fixed fields to each log record
func (l *zapLogger) WithFields(fields LogFields) Logger {
	var f = make([]interface{}, 0)
	for k, v := range fields {
		f = append(f, k)
		f = append(f, v)
	}
	newLogger := l.sugaredLogger.With(f...)
	return &zapLogger{newLogger, l.kafkaWriter}
}

// WithKafkaFilterFn adds a filter function for each kafka record
func (l *zapLogger) WithKafkaFilterFn(filterFn FilterFunc) Logger {
	l.kafkaWriter.kp.config.filterFn = filterFn
	return l
}

// WithKafkaKeyFn adds a key function for each kafka record
func (l *zapLogger) WithKafkaKeyFn(keyFn KeyFunc) Logger {
	l.kafkaWriter.kp.config.keyFn = keyFn
	return l
}
