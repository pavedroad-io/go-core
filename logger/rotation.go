package logger

import (
	"io"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// RotationConfiguration stores the config for log rotation
type RotationConfiguration struct {
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	LocalTime  bool
	Compress   bool
}

func rotationLogger(config RotationConfiguration) io.Writer {

	// use default filename if empty string
	filename := config.Filename
	if filename == "" {
		filename = defaultRotationConfiguration.Filename
	}

	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		LocalTime:  config.LocalTime,
		Compress:   config.Compress,
	}
}
