package lovelock

import (
	"context"

	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inconfirm "github.com/niflaot/pixels/networking/inbound/furniture/lovelock/confirm"
)

// Handler handles explicit second-player lovelock decisions.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings resolves authenticated connections.
	Bindings *binding.Registry
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Service coordinates lovelock state.
	Service *Service
}

// Register adds lovelock confirmation and generic-use behavior.
func Register(registry *netconn.HandlerRegistry, essentials *essential.Service, handler Handler) {
	if essentials != nil {
		essentials.AddExternal(handler.Service)
	}
	if registry != nil {
		_ = registry.Register(inconfirm.Header, handler.Handle)
	}
}

// Handle applies a confirmed or canceled lovelock decision.
func (handler Handler) Handle(connection netconn.Context, packet codec.Packet) error {
	payload, err := inconfirm.Decode(packet)
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
	if !found {
		return nil
	}
	return handler.Service.Confirm(context.Background(), essential.Request{PlayerID: player.ID(), Room: active, Item: item, Target: connection}, payload.Confirmed)
}
