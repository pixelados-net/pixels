package wardrobe

import (
	"context"

	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	"github.com/niflaot/pixels/networking/connection"
	outclothing "github.com/niflaot/pixels/networking/outbound/user/clothing/sets"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// RegisterBootstrap projects clothing unlocks after authentication completes.
func RegisterBootstrap(lifecycle fx.Lifecycle, subscriber bus.Subscriber, service *Service, connections *connection.Registry) error {
	subscription, err := subscriber.Subscribe(playerconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerconnected.Payload)
		if !ok || payload.PlayerID <= 0 || service == nil || connections == nil {
			return nil
		}
		snapshot, loadErr := service.Clothing(ctx, payload.PlayerID)
		if loadErr != nil {
			return loadErr
		}
		packet, encodeErr := outclothing.Encode(snapshot.FigureSetIDs, snapshot.ProductCodes)
		if encodeErr != nil {
			return encodeErr
		}
		target, found := connections.Get(payload.ConnectionKind, payload.ConnectionID)
		if !found {
			return nil
		}
		return target.Send(ctx, packet)
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
