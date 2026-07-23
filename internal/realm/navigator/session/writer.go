package session

import (
	"context"
	"sync"
	"time"

	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	navrecord "github.com/niflaot/pixels/internal/realm/navigator/record"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// PreferenceWriter coalesces repeated Navigator resize persistence by player.
type PreferenceWriter struct {
	// navigator persists final preference values.
	navigator navservice.Manager
	// log records asynchronous persistence failures.
	log *zap.Logger
	// interval controls periodic flush cadence.
	interval time.Duration
	// limit bounds distinct pending players.
	limit int
	// mutex protects pending replacements.
	mutex sync.Mutex
	// pending stores the latest value per player.
	pending map[int64]navrecord.Preference
}

// NewPreferenceWriter creates a bounded coalescing preference writer.
func NewPreferenceWriter(navigator navservice.Manager, log *zap.Logger, interval time.Duration, limit int) *PreferenceWriter {
	if interval <= 0 {
		interval = 250 * time.Millisecond
	}
	if limit <= 0 {
		limit = 4096
	}
	return &PreferenceWriter{navigator: navigator, log: log, interval: interval, limit: limit, pending: make(map[int64]navrecord.Preference)}
}

// RegisterWriter owns the preference writer lifecycle.
func RegisterWriter(lifecycle fx.Lifecycle, writer *PreferenceWriter) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	lifecycle.Append(fx.Hook{OnStart: func(context.Context) error {
		go writer.run(ctx, done)
		return nil
	}, OnStop: func(stopCtx context.Context) error {
		cancel()
		select {
		case <-done:
			return nil
		case <-stopCtx.Done():
			return stopCtx.Err()
		}
	}})
}

// Enqueue replaces one player's pending preference without database I/O.
func (writer *PreferenceWriter) Enqueue(preference navrecord.Preference) bool {
	writer.mutex.Lock()
	defer writer.mutex.Unlock()
	if _, found := writer.pending[preference.PlayerID]; !found && len(writer.pending) >= writer.limit {
		return false
	}
	writer.pending[preference.PlayerID] = preference
	return true
}

// run flushes final values at a bounded cadence and on shutdown.
func (writer *PreferenceWriter) run(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(writer.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			writer.flush(context.Background())
			return
		case <-ticker.C:
			writer.flush(ctx)
		}
	}
}

// flush swaps pending state before performing cold-path writes.
func (writer *PreferenceWriter) flush(ctx context.Context) {
	writer.mutex.Lock()
	pending := writer.pending
	writer.pending = make(map[int64]navrecord.Preference, len(pending))
	writer.mutex.Unlock()
	for playerID, preference := range pending {
		if _, err := writer.navigator.SavePreference(ctx, preference); err != nil && writer.log != nil {
			writer.log.Warn("save navigator preference", zap.Int64("player_id", playerID), zap.Error(err))
		}
	}
}
