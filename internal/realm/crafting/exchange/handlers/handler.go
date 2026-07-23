// Package handlers adapts furniture exchange packets to redemption.
package handlers

import (
	"context"

	craftingexchange "github.com/niflaot/pixels/internal/realm/crafting/exchange"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnitureprojection "github.com/niflaot/pixels/internal/realm/furniture/projection"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inredeem "github.com/niflaot/pixels/networking/inbound/furniture/exchange/redeem"
	outroomremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler handles furniture-for-credit redemption.
type Handler struct {
	Service     *craftingexchange.Service
	Players     *playerlive.Registry
	Bindings    *binding.Registry
	Runtime     *roomlive.Registry
	Connections *netconn.Registry
	// Translations localizes exchange rejections.
	Translations i18n.Translator
}

// New creates an exchange packet handler.
func New(service *craftingexchange.Service, players *playerlive.Registry, bindings *binding.Registry, runtime *roomlive.Registry, connections *netconn.Registry, translations i18n.Translator) *Handler {
	return &Handler{Service: service, Players: players, Bindings: bindings, Runtime: runtime, Connections: connections, Translations: translations}
}

// Handle redeems one owned inventory or same-room credit furniture.
func (handler *Handler) Handle(connection netconn.Context, packet codec.Packet) error {
	payload, err := inredeem.Decode(packet)
	if err != nil {
		return err
	}
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if payload.ItemID <= 0 {
		return codec.ErrInvalidField
	}
	result, err := handler.Service.Redeem(context.Background(), player.ID(), payload.ItemID)
	if err != nil {
		message := "That furniture cannot be exchanged."
		if handler.Translations != nil {
			message = handler.Translations.Default("crafting.error.unavailable")
		}
		response, encodeErr := outalert.Encode(message)
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(context.Background(), response)
	}
	if result.Item.RoomID != nil {
		active, found := handler.Runtime.Find(*result.Item.RoomID)
		if found {
			_, _ = active.ReloadFurniture(result.Item.ID, nil)
			removed, encodeErr := outroomremove.Encode(result.Item.ID, result.Item.OwnerPlayerID)
			if encodeErr != nil {
				return encodeErr
			}
			if err = broadcast.RoomPacket(context.Background(), handler.Connections, active, removed, 0); err != nil {
				return err
			}
		}
	}
	return furnitureprojection.Removed(context.Background(), connection, result.RemovedItemID)
}

// Register adds the exchange header to a connection registry.
func Register(registry *netconn.HandlerRegistry, handler *Handler) {
	_ = registry.Register(inredeem.Header, handler.Handle)
}
