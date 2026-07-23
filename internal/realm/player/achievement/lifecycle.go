package achievement

import (
	"context"

	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// RegisterLifecycle warms and releases online badge snapshots.
func RegisterLifecycle(lifecycle fx.Lifecycle, subscriber bus.Subscriber, service *Service) error {
	connected, err := subscriber.Subscribe(playerconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerconnected.Payload)
		if !ok || payload.PlayerID <= 0 {
			return nil
		}
		return service.Load(ctx, payload.PlayerID)
	})
	if err != nil {
		return err
	}
	disconnected, err := subscriber.Subscribe(playerdisconnected.Name, bus.PriorityNormal, func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerdisconnected.Payload)
		if ok {
			service.Unload(payload.PlayerID)
		}
		return nil
	})
	if err != nil {
		connected.Unsubscribe()
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		connected.Unsubscribe()
		disconnected.Unsubscribe()
		return nil
	}})
	return nil
}
