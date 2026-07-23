package group

import (
	"context"
	"fmt"

	"github.com/niflaot/pixels/internal/realm/group/forum"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	playerauthenticated "github.com/niflaot/pixels/internal/realm/player/events/authenticated"
	sessionunbound "github.com/niflaot/pixels/internal/realm/session/events/unbound"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// RegisterSnapshots hydrates player membership generations and releases session state.
func RegisterSnapshots(lifecycle fx.Lifecycle, subscriber bus.Subscriber, groups *Service, cache *groupruntime.Cache, cursors *forum.Cursors) error {
	authenticated, err := subscriber.Subscribe(playerauthenticated.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerauthenticated.Payload)
		if !ok || payload.PlayerID <= 0 || payload.Reason != "" {
			return nil
		}
		return groups.PreparePlayer(ctx, payload.PlayerID)
	})
	if err != nil {
		return err
	}
	unbound, err := subscriber.Subscribe(sessionunbound.Name, bus.PriorityNormal, func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(sessionunbound.Payload)
		if !ok {
			return nil
		}
		cache.DeletePlayer(payload.Binding.PlayerID)
		cursors.Close(fmt.Sprintf("%s:%s", payload.Binding.ConnectionKind, payload.Binding.ConnectionID))
		return nil
	})
	if err != nil {
		authenticated.Unsubscribe()
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		authenticated.Unsubscribe()
		unbound.Unsubscribe()
		return nil
	}})
	return nil
}
