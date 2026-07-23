// Package handlers adapts recycler packets to its transactional service.
package handlers

import (
	"context"

	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	craftingrecycler "github.com/niflaot/pixels/internal/realm/crafting/recycler"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnitureprojection "github.com/niflaot/pixels/internal/realm/furniture/projection"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inprizes "github.com/niflaot/pixels/networking/inbound/recycler/prizes/get"
	inrecycle "github.com/niflaot/pixels/networking/inbound/recycler/recycle"
	instatus "github.com/niflaot/pixels/networking/inbound/recycler/status/get"
	outprizes "github.com/niflaot/pixels/networking/outbound/recycler/prizes/get"
	outrecycle "github.com/niflaot/pixels/networking/outbound/recycler/recycle"
	outstatus "github.com/niflaot/pixels/networking/outbound/recycler/status/get"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler handles status, recycle, and compatibility prize requests.
type Handler struct {
	Service  *craftingrecycler.Service
	Store    craftingrecord.Store
	Players  *playerlive.Registry
	Bindings *binding.Registry
	// Translations localizes recycler rejections.
	Translations i18n.Translator
}

// New creates a grouped recycler packet handler.
func New(service *craftingrecycler.Service, store craftingrecord.Store, players *playerlive.Registry, bindings *binding.Registry, translations i18n.Translator) *Handler {
	return &Handler{Service: service, Store: store, Players: players, Bindings: bindings, Translations: translations}
}

// Handle decodes and executes one recycler request.
func (handler *Handler) Handle(connection netconn.Context, packet codec.Packet) error {
	ctx := context.Background()
	switch packet.Header {
	case instatus.Header:
		if err := instatus.Decode(packet); err != nil {
			return err
		}
		status := outstatus.Disabled
		if handler.Service.Config().RecyclerEnabled {
			status = outstatus.Enabled
		}
		response, err := outstatus.Encode(status, 0)
		if err != nil {
			return err
		}
		return connection.Send(ctx, response)
	case inprizes.Header:
		if err := inprizes.Decode(packet); err != nil {
			return err
		}
		prizes, err := handler.Store.Prizes(ctx)
		if err != nil {
			return err
		}
		var chances [6]int32
		for tier, chance := range handler.Service.Config().RecyclerRarityChance {
			chances[tier] = int32(chance)
		}
		chances[1] = 1
		response, err := outprizes.Encode(prizes, chances)
		if err != nil {
			return err
		}
		return connection.Send(ctx, response)
	case inrecycle.Header:
		payload, err := inrecycle.Decode(packet)
		if err != nil {
			return err
		}
		player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
		if err != nil {
			return err
		}
		result, err := handler.Service.Recycle(ctx, player.ID(), payload.ItemIDs)
		if err != nil {
			if err != craftingrecord.ErrRecyclerClosed {
				message := "The selected furniture cannot be recycled."
				if handler.Translations != nil {
					message = handler.Translations.Default("crafting.error.ingredients")
				}
				response, encodeErr := outalert.Encode(message)
				if encodeErr != nil {
					return encodeErr
				}
				return connection.Send(ctx, response)
			}
			response, encodeErr := outrecycle.Encode(outrecycle.Closed, 0)
			if encodeErr != nil {
				return encodeErr
			}
			return connection.Send(ctx, response)
		}
		if err = furnitureprojection.Inventory(ctx, connection, result.Removed, result.Granted, result.Definition); err != nil {
			return err
		}
		response, err := outrecycle.Encode(outrecycle.Complete, result.Prize.RewardDefinitionID)
		if err != nil {
			return err
		}
		return connection.Send(ctx, response)
	default:
		return codec.ErrUnexpectedHeader
	}
}

// Register adds every recycler header to a connection registry.
func Register(registry *netconn.HandlerRegistry, handler *Handler) {
	for _, header := range []uint16{instatus.Header, inprizes.Header, inrecycle.Header} {
		_ = registry.Register(header, handler.Handle)
	}
}
