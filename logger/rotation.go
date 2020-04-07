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
	rotConfig := DefaultRotationCfg()
	// override rotation defaults with config
	if config.Filename != "" {
		rotConfig.Filename = config.Filename
	}
	if config.MaxSize >= 0 {
		rotConfig.MaxSize = config.MaxSize
	}
	if config.MaxAge >= 0 {
		rotConfig.MaxAge = config.MaxAge
	}
	if config.MaxBackups >= 0 {
		rotConfig.MaxBackups = config.MaxBackups
	}
	rotConfig.LocalTime = config.LocalTime
	rotConfig.Compress = config.Compress

	return &lumberjack.Logger{
		Filename:   rotConfig.Filename,
		MaxSize:    rotConfig.MaxSize,
		MaxBackups: rotConfig.MaxBackups,
		MaxAge:     rotConfig.MaxAge,
		LocalTime:  rotConfig.LocalTime,
		Compress:   rotConfig.Compress,
	}
}
