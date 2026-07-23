// Package commerce adapts Marketplace purchase and proceeds packets.
package commerce

import (
	"context"
	"errors"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbuy "github.com/niflaot/pixels/networking/inbound/marketplace/buy"
	inredeem "github.com/niflaot/pixels/networking/inbound/marketplace/redeem"
	intokens "github.com/niflaot/pixels/networking/inbound/marketplace/tokens"
	outbuyresult "github.com/niflaot/pixels/networking/outbound/marketplace/buyresult"
	outown "github.com/niflaot/pixels/networking/outbound/marketplace/own"
	outtokens "github.com/niflaot/pixels/networking/outbound/marketplace/tokens"
)

// Handler owns Marketplace commerce adapters.
type Handler struct {
	// Service executes Marketplace commerce behavior.
	Service *marketcore.Service
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
}

// Register installs Marketplace commerce handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(inbuy.Header, handler.buy)
	_ = registry.Register(intokens.Header, handler.tokens)
	_ = registry.Register(inredeem.Header, handler.redeem)
}

// playerID resolves a bound player.
func (handler Handler) playerID(connection netconn.Context) (int64, error) {
	bound, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	if !found {
		return 0, binding.ErrBindingNotFound
	}
	return bound.PlayerID, nil
}

// buy settles one listing.
func (handler Handler) buy(connection netconn.Context, packet codec.Packet) error {
	offerID, err := inbuy.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.playerID(connection)
	if err != nil {
		return err
	}
	listing, serviceErr := handler.Service.Buy(context.Background(), playerID, int64(offerID))
	result := int32(1)
	var replacementID, replacementPrice int64
	if serviceErr != nil {
		switch {
		case errors.Is(serviceErr, marketcore.ErrListingUnavailable):
			result = 2
			replacement, found, replacementErr := handler.Service.Replacement(context.Background(), listing)
			if replacementErr != nil {
				return replacementErr
			}
			if found {
				result = 3
				replacementID = replacement.ID
				replacementPrice = handler.Service.BuyerPrice(replacement.RawPrice)
			}
		case errors.Is(serviceErr, currencyservice.ErrInsufficientBalance):
			result = 4
		case errors.Is(serviceErr, marketcore.ErrOwnListing):
			result = 2
		default:
			return serviceErr
		}
	}
	response, err := outbuyresult.Encode(result, replacementID, replacementPrice, int64(offerID))
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// tokens purchases the configured token package.
func (handler Handler) tokens(connection netconn.Context, packet codec.Packet) error {
	if err := intokens.Decode(packet); err != nil {
		return err
	}
	playerID, err := handler.playerID(connection)
	if err != nil {
		return err
	}
	balance, serviceErr := handler.Service.BuyTokens(context.Background(), playerID)
	result := int32(1)
	if serviceErr != nil {
		result = 0
	}
	response, err := outtokens.Encode(result, balance)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// redeem claims seller proceeds.
func (handler Handler) redeem(connection netconn.Context, packet codec.Packet) error {
	if err := inredeem.Decode(packet); err != nil {
		return err
	}
	playerID, err := handler.playerID(connection)
	if err != nil {
		return err
	}
	_, _, serviceErr := handler.Service.Redeem(context.Background(), playerID)
	if serviceErr != nil {
		return serviceErr
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
