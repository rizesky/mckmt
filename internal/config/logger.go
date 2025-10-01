package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger creates a zap logger based on the provided LoggingConfig
func InitLogger(cfg LoggingConfig) (*zap.Logger, error) {
	var config zap.Config

	if cfg.Format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.Level = zap.NewAtomicLevelAt(getLogLevel(cfg.Level))
	config.DisableCaller = !cfg.Caller
	config.DisableStacktrace = !cfg.Stacktrace

	// Set output paths if specified
	if len(cfg.OutputPaths) > 0 {
		config.OutputPaths = cfg.OutputPaths
	}
	if len(cfg.ErrorOutputPaths) > 0 {
		config.ErrorOutputPaths = cfg.ErrorOutputPaths
	}

	return config.Build()
}

// getLogLevel converts string level to zapcore.Level
func getLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
