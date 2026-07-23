package mysterybox

import (
	"context"

	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incancel "github.com/niflaot/pixels/networking/inbound/furniture/mysterybox/cancel"
	introphy "github.com/niflaot/pixels/networking/inbound/furniture/mysterybox/trophy"
	outkeys "github.com/niflaot/pixels/networking/outbound/furniture/mysterybox/keys"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// Handler handles cancellation and trophy inscription requests.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings resolves authenticated connections.
	Bindings *binding.Registry
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Connections resolves login bootstrap targets.
	Connections *netconn.Registry
	// Service coordinates mystery-box behavior.
	Service *Service
}

// Register adds mystery-box protocol adapters and generic furniture behavior.
func Register(registry *netconn.HandlerRegistry, essentials *essential.Service, handler Handler) {
	if essentials != nil {
		essentials.AddExternal(handler.Service)
	}
	if registry == nil {
		return
	}
	_ = registry.Register(incancel.Header, handler.cancel)
	_ = registry.Register(introphy.Header, handler.trophy)
}

// RegisterBootstrap sends key tracker state after authentication.
func RegisterBootstrap(lifecycle fx.Lifecycle, subscriber bus.Subscriber, handler Handler) error {
	subscription, err := subscriber.Subscribe(playerconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerconnected.Payload)
		if !ok || payload.PlayerID <= 0 {
			return nil
		}
		connection, found := handler.Connections.Get(payload.ConnectionKind, payload.ConnectionID)
		if !found {
			return nil
		}
		keys, findErr := handler.Service.Keys(ctx, payload.PlayerID)
		if findErr != nil {
			return findErr
		}
		packet, encodeErr := outkeys.Encode(keys.BoxColor, keys.KeyColor)
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, packet)
	})
	if err != nil {
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { subscription.Unsubscribe(); return nil }})
	return nil
}

// cancel invalidates a pending reveal for the authenticated player.
func (handler Handler) cancel(connection netconn.Context, packet codec.Packet) error {
	_, err := incancel.Decode(packet)
	if err != nil {
		return err
	}
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	return handler.Service.Cancel(context.Background(), player.ID(), connection)
}

// trophy stores an owner-authored permanent inscription.
func (handler Handler) trophy(connection netconn.Context, packet codec.Packet) error {
	payload, err := introphy.Decode(packet)
	if err != nil {
		return err
	}
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return nil
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return nil
	}
	item, found := active.FurnitureItem(int64(payload.ItemID))
	if !found || item.OwnerPlayerID != player.ID() || item.Definition.InteractionType != "mystery_trophy" {
		return nil
	}
	encoded, err := handler.Service.Inscribe(context.Background(), player.ID(), roomID, item.ID, item.ExtraData, payload.Text)
	if err != nil {
		return err
	}
	active.SetFurnitureExtraData(item.ID, encoded)
	return nil
}
