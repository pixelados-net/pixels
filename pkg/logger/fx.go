package logger

import (
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewFx creates an Fx event logger backed by zap.
func NewFx(log *zap.Logger) fxevent.Logger {
	logger := &fxevent.ZapLogger{
		Logger: log,
	}
	logger.UseLogLevel(zapcore.DebugLevel)

	return logger
}
