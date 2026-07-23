package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a zap logger from logger configuration.
func New(config Config) (*zap.Logger, error) {
	zapConfig, err := buildConfig(config)
	if err != nil {
		return nil, err
	}

	return zapConfig.Build()
}

// buildConfig creates a zap configuration from logger settings.
func buildConfig(config Config) (zap.Config, error) {
	var level zapcore.Level

	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return zap.Config{}, fmt.Errorf("parse log level: %w", err)
	}

	if config.Format != FormatConsole && config.Format != FormatJSON {
		return zap.Config{}, fmt.Errorf("unsupported log format: %s", config.Format)
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.Encoding = string(config.Format)
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if config.Format == FormatConsole {
		zapConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	if config.ToonConsole {
		applyToonConsole(&zapConfig)
	}

	return zapConfig, nil
}

// applyToonConsole trims zap output for protocol tracing.
func applyToonConsole(config *zap.Config) {
	config.Encoding = FormatTOON
	config.DisableCaller = true
	config.DisableStacktrace = true
	config.EncoderConfig.TimeKey = ""
	config.EncoderConfig.CallerKey = ""
	config.EncoderConfig.LevelKey = "lvl"
	config.EncoderConfig.MessageKey = "msg"
	config.EncoderConfig.EncodeLevel = toonLevelEncoder
}
