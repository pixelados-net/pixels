package broadcast

import (
	"context"

	rightsgranted "github.com/niflaot/pixels/internal/realm/room/events/rightsgranted"
	rightsrevoked "github.com/niflaot/pixels/internal/realm/room/events/rightsrevoked"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Register subscribes the room rights runtime projection.
func Register(lifecycle fx.Lifecycle, subscriber bus.Subscriber, broadcaster *Broadcaster, log *zap.Logger) error {
	granted, err := subscriber.Subscribe(rightsgranted.Name, bus.PriorityLow, deferGranted(broadcaster, log))
	if err != nil {
		return err
	}
	revoked, err := subscriber.Subscribe(rightsrevoked.Name, bus.PriorityLow, deferRevoked(broadcaster, log))
	if err != nil {
		granted.Unsubscribe()
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		granted.Unsubscribe()
		revoked.Unsubscribe()
		return nil
	}})

	return nil
}

// deferGranted defers grant projection until transaction commit.
func deferGranted(broadcaster *Broadcaster, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(rightsgranted.Payload)
		if !ok {
			return nil
		}
		project := func(callbackCtx context.Context) {
			if err := broadcaster.Granted(callbackCtx, payload.RoomID, payload.PlayerID); err != nil {
				log.Warn("room rights grant projection failed", zap.Error(err), zap.Int64("room_id", payload.RoomID), zap.Int64("player_id", payload.PlayerID))
			}
		}
		if !postgres.AfterCommit(ctx, project) {
			project(ctx)
		}

		return nil
	}
}

// deferRevoked defers revocation projection until transaction commit.
func deferRevoked(broadcaster *Broadcaster, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(rightsrevoked.Payload)
		if !ok {
			return nil
		}
		project := func(callbackCtx context.Context) {
			if err := broadcaster.Revoked(callbackCtx, payload.RoomID, payload.PlayerID); err != nil {
				log.Warn("room rights revocation projection failed", zap.Error(err), zap.Int64("room_id", payload.RoomID), zap.Int64("player_id", payload.PlayerID))
			}
		}
		if !postgres.AfterCommit(ctx, project) {
			project(ctx)
		}

		return nil
	}
}
