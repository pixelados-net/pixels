package effect

import (
	"context"

	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inactivate "github.com/niflaot/pixels/networking/inbound/user/effect/activate"
	inenable "github.com/niflaot/pixels/networking/inbound/user/effect/enable"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Handler adapts Nitro effect requests to the effect service.
type Handler struct {
	// Effects executes effect behavior.
	Effects *Service
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
	// Log records non-fatal handler failures.
	Log *zap.Logger
}

// RegisterBootstrap sends the effect inventory after authentication completes.
func RegisterBootstrap(lifecycle fx.Lifecycle, subscriber bus.Subscriber, service *Service) error {
	subscription, err := subscriber.Subscribe(playerconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerconnected.Payload)
		if !ok || payload.PlayerID <= 0 {
			return nil
		}
		return service.SendInventory(ctx, payload.PlayerID)
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

// RegisterHandlers registers effect activation and selection packets.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(inenable.Header, handler.enable)
	_ = registry.Register(inactivate.Header, handler.activate)
}

// enable selects one player effect.
func (handler Handler) enable(connection netconn.Context, packet codec.Packet) error {
	effectID, err := inenable.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Effects == nil {
		return nil
	}
	return handler.Effects.Enable(context.Background(), playerID, effectID)
}

// activate starts one player effect charge.
func (handler Handler) activate(connection netconn.Context, packet codec.Packet) error {
	effectID, err := inactivate.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Effects == nil {
		return nil
	}
	_, err = handler.Effects.Activate(context.Background(), playerID, effectID)
	return err
}

// playerID resolves an authenticated connection binding.
func (handler Handler) playerID(connection netconn.Context) (int64, bool) {
	if handler.Bindings == nil {
		return 0, false
	}
	current, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	return current.PlayerID, found
}
