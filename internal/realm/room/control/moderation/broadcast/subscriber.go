package broadcast

import (
	"context"

	bannedevent "github.com/niflaot/pixels/internal/realm/room/control/events/banned"
	kickedevent "github.com/niflaot/pixels/internal/realm/room/control/events/kicked"
	mutedevent "github.com/niflaot/pixels/internal/realm/room/control/events/muted"
	unbannedevent "github.com/niflaot/pixels/internal/realm/room/control/events/unbanned"
	unmutedevent "github.com/niflaot/pixels/internal/realm/room/control/events/unmuted"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// eventHandler pairs one event name with its projection handler.
type eventHandler struct {
	// name identifies the subscribed event.
	name bus.Name
	// handler projects the subscribed event.
	handler bus.Handler
}

// Register subscribes committed moderation runtime projections.
func Register(lifecycle fx.Lifecycle, subscriber bus.Subscriber, broadcaster *Broadcaster, log *zap.Logger) error {
	handlers := []eventHandler{
		{name: kickedevent.Name, handler: handleKick(broadcaster, log)},
		{name: mutedevent.Name, handler: handleMute(broadcaster, log)},
		{name: unmutedevent.Name, handler: handleUnmute(broadcaster, log)},
		{name: bannedevent.Name, handler: handleBan(broadcaster, log)},
		{name: unbannedevent.Name, handler: handleUnban(broadcaster, log)},
	}
	subscriptions := make([]*bus.Subscription, 0, len(handlers))
	for _, item := range handlers {
		subscription, err := subscriber.Subscribe(item.name, bus.PriorityLow, item.handler)
		if err != nil {
			unsubscribeAll(subscriptions)
			return err
		}
		subscriptions = append(subscriptions, subscription)
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		unsubscribeAll(subscriptions)
		return nil
	}})

	return nil
}

// handleKick projects a committed kick.
func handleKick(broadcaster *Broadcaster, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(kickedevent.Payload)
		if ok {
			deferProjection(ctx, log, payload.RoomID, payload.TargetPlayerID, func(callbackCtx context.Context) error {
				return broadcaster.Kick(callbackCtx, payload.RoomID, payload.TargetPlayerID)
			})
		}
		return nil
	}
}

// handleMute projects a committed mute.
func handleMute(broadcaster *Broadcaster, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(mutedevent.Payload)
		if ok {
			deferProjection(ctx, log, payload.RoomID, payload.TargetPlayerID, func(callbackCtx context.Context) error {
				return broadcaster.Mute(callbackCtx, payload.RoomID, payload.TargetPlayerID, payload.DurationSeconds)
			})
		}
		return nil
	}
}

// handleUnmute projects a committed unmute.
func handleUnmute(broadcaster *Broadcaster, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(unmutedevent.Payload)
		if ok {
			deferProjection(ctx, log, payload.RoomID, payload.TargetPlayerID, func(callbackCtx context.Context) error {
				return broadcaster.Mute(callbackCtx, payload.RoomID, payload.TargetPlayerID, 0)
			})
		}
		return nil
	}
}

// handleBan projects a committed ban.
func handleBan(broadcaster *Broadcaster, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(bannedevent.Payload)
		if ok {
			deferProjection(ctx, log, payload.RoomID, payload.TargetPlayerID, func(callbackCtx context.Context) error {
				return broadcaster.Ban(callbackCtx, payload.RoomID, payload.TargetPlayerID)
			})
		}
		return nil
	}
}

// handleUnban projects a committed unban.
func handleUnban(broadcaster *Broadcaster, log *zap.Logger) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(unbannedevent.Payload)
		if ok {
			deferProjection(ctx, log, payload.RoomID, payload.TargetPlayerID, func(callbackCtx context.Context) error {
				return broadcaster.Unban(callbackCtx, payload.RoomID, payload.TargetPlayerID)
			})
		}
		return nil
	}
}

// deferProjection executes a runtime projection only after a successful commit.
func deferProjection(ctx context.Context, log *zap.Logger, roomID int64, playerID int64, projection func(context.Context) error) {
	callback := func(callbackCtx context.Context) {
		if err := projection(callbackCtx); err != nil {
			log.Warn("room moderation projection failed", zap.Error(err), zap.Int64("room_id", roomID), zap.Int64("player_id", playerID))
		}
	}
	if !postgres.AfterCommit(ctx, callback) {
		callback(ctx)
	}
}

// unsubscribeAll removes moderation projection subscriptions.
func unsubscribeAll(subscriptions []*bus.Subscription) {
	for _, subscription := range subscriptions {
		subscription.Unsubscribe()
	}
}
