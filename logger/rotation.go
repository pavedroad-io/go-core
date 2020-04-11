package logger

import (
	"io"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// RotationConfiguration stores the config for log rotation
type RotationConfiguration struct {
	MaxSize    int
	MaxAge     int
	MaxBackups int
	LocalTime  bool
	Compress   bool
}

func rotationLogger(filename string, config RotationConfiguration) io.Writer {

	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		LocalTime:  config.LocalTime,
		Compress:   config.Compress,
	}
}
