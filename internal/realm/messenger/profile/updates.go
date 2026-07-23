package profile

import (
	"context"

	messengercore "github.com/niflaot/pixels/internal/realm/messenger/core"
	"github.com/niflaot/pixels/internal/realm/messenger/runtime/delivery"
	namechanged "github.com/niflaot/pixels/internal/realm/player/identity/events/namechanged"
	profileupdated "github.com/niflaot/pixels/internal/realm/player/profile/events/updated"
	outchanged "github.com/niflaot/pixels/networking/outbound/user/profile/changed"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// UpdateBroadcaster refreshes friend cards and bounded public-profile observers.
type UpdateBroadcaster struct {
	// messenger invalidates durable public cards and resolves observers.
	messenger *messengercore.Service
	// presence refreshes online friend cards.
	presence *PresenceBroadcaster
	// delivery sends targeted profile invalidations.
	delivery *delivery.Sender
}

// NewUpdates creates public-profile mutation projection behavior.
func NewUpdates(messenger *messengercore.Service, presence *PresenceBroadcaster, delivery *delivery.Sender) *UpdateBroadcaster {
	return &UpdateBroadcaster{messenger: messenger, presence: presence, delivery: delivery}
}

// Refresh invalidates cards, refreshes online friends, and notifies active observers.
func (broadcaster *UpdateBroadcaster) Refresh(ctx context.Context, playerID int64) error {
	broadcaster.messenger.InvalidateProfile(ctx, playerID)
	if err := broadcaster.presence.Update(ctx, playerID); err != nil {
		return err
	}
	packet, err := outchanged.Encode(int32(playerID))
	if err != nil {
		return err
	}
	for _, viewerID := range broadcaster.messenger.RelationshipViewers(playerID) {
		if _, err = broadcaster.delivery.Send(ctx, viewerID, packet); err != nil {
			return err
		}
	}
	return nil
}

// RegisterUpdates subscribes committed identity and profile mutations.
func RegisterUpdates(lifecycle fx.Lifecycle, subscriber bus.Subscriber, broadcaster *UpdateBroadcaster, log *zap.Logger) error {
	handler := func(playerID func(any) (int64, bool)) bus.Handler {
		return func(ctx context.Context, event bus.Event) error {
			id, ok := playerID(event.Payload)
			if !ok {
				return nil
			}
			if err := broadcaster.Refresh(ctx, id); err != nil && log != nil {
				log.Warn("messenger public profile refresh failed", zap.Int64("player_id", id), zap.Error(err))
			}
			return nil
		}
	}
	subscriptions := make([]*bus.Subscription, 0, 2)
	first, err := subscriber.Subscribe(namechanged.Name, bus.PriorityLow, handler(nameChangedID))
	if err != nil {
		return err
	}
	subscriptions = append(subscriptions, first)
	second, err := subscriber.Subscribe(profileupdated.Name, bus.PriorityLow, handler(profileUpdatedID))
	if err != nil {
		unsubscribe(subscriptions)
		return err
	}
	subscriptions = append(subscriptions, second)
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		unsubscribe(subscriptions)
		return nil
	}})
	return nil
}

// nameChangedID extracts an identity change player identifier.
func nameChangedID(payload any) (int64, bool) {
	value, ok := payload.(namechanged.Payload)
	return value.PlayerID, ok
}

// profileUpdatedID extracts a profile change player identifier.
func profileUpdatedID(payload any) (int64, bool) {
	value, ok := payload.(profileupdated.Payload)
	return value.PlayerID, ok
}
