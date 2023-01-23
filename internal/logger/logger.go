// Package logger contains functions for a working with application logging.
package logger

import (
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates new "zap" logger with a small customization.
func New(l Level, f Format) (*zap.Logger, error) {
	var config zap.Config

	switch f {
	case ConsoleFormat:
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")

	case JSONFormat:
		config = zap.NewProductionConfig() // json encoder is used by default

	default:
		return nil, errors.New("unsupported logging format")
	}

	// default configuration for all encoders
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.Development = false
	config.DisableStacktrace = true
	config.DisableCaller = true

	// enable additional features for debugging
	if l <= DebugLevel {
		config.Development = true
		config.DisableStacktrace = false
		config.DisableCaller = false
	}

	var zapLvl zapcore.Level

	switch l { // convert level to zap.Level
	case DebugLevel:
		zapLvl = zap.DebugLevel
	case InfoLevel:
		zapLvl = zap.InfoLevel
	case WarnLevel:
		zapLvl = zap.WarnLevel
	case ErrorLevel:
		zapLvl = zap.ErrorLevel
	case FatalLevel:
		zapLvl = zap.FatalLevel
	default:
		return nil, errors.New("unsupported logging level")
	}

	config.Level = zap.NewAtomicLevelAt(zapLvl)

	return config.Build()
}
