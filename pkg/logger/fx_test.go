package logger

import (
	"testing"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestNewFxLogsEventsAtDebug verifies Fx lifecycle noise only appears at debug level.
func TestNewFxLogsEventsAtDebug(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	logger := NewFx(zap.New(core))

	logger.LogEvent(&fxevent.Started{})

	if logs.Len() != 0 {
		t.Fatalf("expected no info-level Fx event logs, got %d", logs.Len())
	}
}

// TestNewFxUsesZapLogger verifies Fx events are emitted through zap at debug level.
func TestNewFxUsesZapLogger(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	logger := NewFx(zap.New(core))

	logger.LogEvent(&fxevent.Started{})

	if logs.Len() != 1 {
		t.Fatalf("expected one debug Fx event log, got %d", logs.Len())
	}
}
