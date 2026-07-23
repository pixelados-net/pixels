// Package listing adapts Marketplace listing lifecycle packets.
package listing

import (
	"context"
	"errors"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incancel "github.com/niflaot/pixels/networking/inbound/marketplace/cancel"
	incansell "github.com/niflaot/pixels/networking/inbound/marketplace/cansell"
	inown "github.com/niflaot/pixels/networking/inbound/marketplace/own"
	insell "github.com/niflaot/pixels/networking/inbound/marketplace/sell"
	outcancel "github.com/niflaot/pixels/networking/outbound/marketplace/cancel"
	outown "github.com/niflaot/pixels/networking/outbound/marketplace/own"
	outposted "github.com/niflaot/pixels/networking/outbound/marketplace/posted"
	outtokens "github.com/niflaot/pixels/networking/outbound/marketplace/tokens"
)

// Handler owns Marketplace listing adapters.
type Handler struct {
	// Service executes Marketplace listing behavior.
	Service *marketcore.Service
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
}

// Register installs Marketplace listing handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(incansell.Header, handler.canSell)
	_ = registry.Register(insell.Header, handler.sell)
	_ = registry.Register(incancel.Header, handler.cancel)
	_ = registry.Register(inown.Header, handler.own)
}

// playerID resolves a bound player.
func (handler Handler) playerID(connection netconn.Context) (int64, error) {
	bound, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	if !found {
		return 0, binding.ErrBindingNotFound
	}
	return bound.PlayerID, nil
}

// canSell reports eligibility and token balance.
func (handler Handler) canSell(connection netconn.Context, packet codec.Packet) error {
	if err := incansell.Decode(packet); err != nil {
		return err
	}
	playerID, err := handler.playerID(connection)
	if err != nil {
		return err
	}
	count, err := handler.Service.Tokens(context.Background(), playerID)
	if err != nil {
		return err
	}
	result := int32(0)
	if handler.Service.CanSell() {
		result = 1
	}
	response, err := outtokens.Encode(result, count)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// sell creates one listing.
func (handler Handler) sell(connection netconn.Context, packet codec.Packet) error {
	payload, err := insell.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.playerID(connection)
	if err != nil {
		return err
	}
	_, serviceErr := handler.Service.List(context.Background(), playerID, int64(payload.ItemID), int64(payload.Price))
	result := int32(1)
	if serviceErr != nil {
		switch {
		case errors.Is(serviceErr, marketcore.ErrDisabled):
			result = 3
		case errors.Is(serviceErr, marketcore.ErrNoToken), errors.Is(serviceErr, marketcore.ErrInvalidPrice), errors.Is(serviceErr, furnitureservice.ErrItemUnavailable):
			result = 2
		default:
			return serviceErr
		}
	}
	response, err := outposted.Encode(result)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// cancel closes one seller listing.
func (handler Handler) cancel(connection netconn.Context, packet codec.Packet) error {
	offerID, err := incancel.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.playerID(connection)
	if err != nil {
		return err
	}
	serviceErr := handler.Service.Close(context.Background(), int64(offerID), playerID, false)
	response, err := outcancel.Encode(int64(offerID), serviceErr == nil)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// own sends seller listings and waiting credits.
func (handler Handler) own(connection netconn.Context, packet codec.Packet) error {
	if err := inown.Decode(packet); err != nil {
		return err
	}
	playerID, err := handler.playerID(connection)
	if err != nil {
		return err
	}
	offers, waiting, err := handler.Service.OwnListings(context.Background(), playerID)
	if err != nil {
		return err
	}
	response, err := outown.Encode(waiting, offers)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}
