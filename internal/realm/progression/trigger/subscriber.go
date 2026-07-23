// Package trigger maps committed realm events into progression counters.
package trigger

import (
	"context"
	"errors"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// Subscriber maps committed gameplay events into bounded progression work.
type Subscriber struct {
	// engine receives normalized trigger deltas.
	engine Progressor
	// players resolves live registration metadata.
	players *playerlive.Registry
	// log records deferred progression failures.
	log *zap.Logger
}

// Progressor receives normalized gameplay progression from cross-realm events.
type Progressor interface {
	// Progress queues one ordinary trigger.
	Progress(context.Context, int64, string, int64) error
	// ProgressData queues one metadata-bearing trigger.
	ProgressData(context.Context, int64, string, string, int64) error
	// ProgressDaily queues one UTC-day-idempotent trigger.
	ProgressDaily(context.Context, int64, string, int64) error
	// HydratePlayer loads durable achievement positions before queued work.
	HydratePlayer(context.Context, int64) error
	// SetTriggerProgress sets absolute progress for matching achievements.
	SetTriggerProgress(context.Context, int64, string, int64) error
	// FlushPlayer persists one player's pending deltas.
	FlushPlayer(context.Context, int64) error
	// ForgetPlayer releases one disconnected player's forecast state.
	ForgetPlayer(int64)
}

// registration binds one event name to one adapter.
type registration struct {
	// name identifies the event.
	name bus.Name
	// handler adapts its payload.
	handler bus.Handler
}

// New creates one cross-realm progression subscriber.
func New(engine *progressionengine.Service, players *playerlive.Registry, log *zap.Logger) *Subscriber {
	if log == nil {
		log = zap.NewNop()
	}
	return &Subscriber{engine: engine, players: players, log: log}
}

// Register installs every supported gameplay trigger.
func Register(subscriber bus.Subscriber, adapter *Subscriber) error {
	var result error
	for _, value := range adapter.registrations() {
		_, err := subscriber.Subscribe(value.name, bus.PriorityNormal, value.handler)
		result = errors.Join(result, err)
	}
	return result
}

// progress queues one trigger without leaking progression failures into gameplay.
func (subscriber *Subscriber) progress(ctx context.Context, playerID int64, key string, amount int64, daily bool) {
	var err error
	if daily {
		err = subscriber.engine.ProgressDaily(ctx, playerID, key, amount)
	} else {
		err = subscriber.engine.Progress(ctx, playerID, key, amount)
	}
	if err != nil {
		subscriber.log.Warn("progression trigger failed", zap.Int64("player_id", playerID), zap.String("trigger", key), zap.Error(err))
	}
}

// progressData queues one metadata-bearing trigger without leaking failures into gameplay.
func (subscriber *Subscriber) progressData(ctx context.Context, playerID int64, key string, data string, amount int64) {
	if err := subscriber.engine.ProgressData(ctx, playerID, key, data, amount); err != nil {
		subscriber.log.Warn("progression trigger failed", zap.Int64("player_id", playerID), zap.String("trigger", key), zap.Error(err))
	}
}
