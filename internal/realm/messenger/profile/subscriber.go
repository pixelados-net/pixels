package profile

import (
	"context"

	relationchanged "github.com/niflaot/pixels/internal/realm/messenger/friend/events/relationchanged"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// Register subscribes profile refreshes to relationship changes.
func RegisterRelationships(lifecycle fx.Lifecycle, subscriber bus.Subscriber, broadcaster *RelationshipBroadcaster) error {
	subscription, err := subscriber.Subscribe(relationchanged.Name, bus.PriorityLow, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(relationchanged.Payload)
		if !ok {
			return nil
		}
		return broadcaster.Refresh(ctx, payload.PlayerID)
	})
	if err != nil {
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		subscription.Unsubscribe()
		return nil
	}})
	return nil
}
