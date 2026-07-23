package profile

import (
	"context"

	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	roomleft "github.com/niflaot/pixels/internal/realm/room/access/events/left"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Register subscribes messenger presence projection to player and room lifecycle events.
func RegisterPresence(lifecycle fx.Lifecycle, subscriber bus.Subscriber, broadcaster *PresenceBroadcaster, log *zap.Logger) error {
	subscriptions := make([]*bus.Subscription, 0, 4)
	registrations := []registration{
		{name: playerconnected.Name, playerID: connectedPlayerID},
		{name: playerdisconnected.Name, playerID: disconnectedPlayerID},
		{name: roomentered.Name, playerID: enteredPlayerID},
		{name: roomleft.Name, playerID: leftPlayerID},
	}
	for _, registration := range registrations {
		subscription, err := subscriber.Subscribe(registration.name, bus.PriorityLow, presenceHandler(broadcaster, log, registration.playerID))
		if err != nil {
			unsubscribe(subscriptions)
			return err
		}
		subscriptions = append(subscriptions, subscription)
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		unsubscribe(subscriptions)
		return nil
	}})
	return nil
}

// registration pairs one lifecycle event with its player-id extractor.
type registration struct {
	// name identifies one lifecycle event.
	name bus.Name
	// playerID extracts its player identifier.
	playerID func(any) (int64, bool)
}

// payloadPlayerID extracts a player identifier from one event payload.
type payloadPlayerID func(any) (int64, bool)

// presenceHandler projects one supported lifecycle event.
func presenceHandler(broadcaster *PresenceBroadcaster, log *zap.Logger, extract payloadPlayerID) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		playerID, ok := extract(event.Payload)
		if !ok {
			return nil
		}
		if err := broadcaster.Update(ctx, playerID); err != nil && log != nil {
			log.Warn("messenger presence projection failed", zap.Int64("player_id", playerID), zap.Error(err))
		}
		return nil
	}
}

// connectedPlayerID extracts a connected player id.
func connectedPlayerID(payload any) (int64, bool) {
	value, ok := payload.(playerconnected.Payload)
	return value.PlayerID, ok
}

// disconnectedPlayerID extracts a disconnected player id.
func disconnectedPlayerID(payload any) (int64, bool) {
	value, ok := payload.(playerdisconnected.Payload)
	return value.PlayerID, ok
}

// enteredPlayerID extracts an entered player id.
func enteredPlayerID(payload any) (int64, bool) {
	value, ok := payload.(roomentered.Payload)
	return value.PlayerID, ok
}

// leftPlayerID extracts a left player id.
func leftPlayerID(payload any) (int64, bool) {
	value, ok := payload.(roomleft.Payload)
	return value.PlayerID, ok
}

// unsubscribe releases presence subscriptions.
func unsubscribe(subscriptions []*bus.Subscription) {
	for _, subscription := range subscriptions {
		subscription.Unsubscribe()
	}
}
