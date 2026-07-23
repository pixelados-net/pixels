package rentable

import (
	"context"

	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incancel "github.com/niflaot/pixels/networking/inbound/furniture/rentable/cancel"
	inextend "github.com/niflaot/pixels/networking/inbound/furniture/rentable/extend"
	inextendstrip "github.com/niflaot/pixels/networking/inbound/furniture/rentable/extendstrip"
	inoffer "github.com/niflaot/pixels/networking/inbound/furniture/rentable/offer"
	inrent "github.com/niflaot/pixels/networking/inbound/furniture/rentable/rent"
	instatus "github.com/niflaot/pixels/networking/inbound/furniture/rentable/status"
	outfailed "github.com/niflaot/pixels/networking/outbound/furniture/rentable/rentfailed"
	outok "github.com/niflaot/pixels/networking/outbound/furniture/rentable/rentok"
	outstatus "github.com/niflaot/pixels/networking/outbound/furniture/rentable/status"
	oukoffer "github.com/niflaot/pixels/networking/outbound/subscription/rentableoffer"
)

const (
	// FailureUnavailable maps guarded rental conflicts to Nitro.
	FailureUnavailable int32 = 100
)

// Handler adapts rentable protocol requests to the service.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
	// Service coordinates rental behavior.
	Service *Service
}

// Register adds every rentable furniture request.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	for _, header := range []uint16{instatus.Header, inrent.Header, incancel.Header, inextend.Header, inextendstrip.Header, inoffer.Header} {
		_ = registry.Register(header, handler.Handle)
	}
}

// Handle routes one rentable request.
func (handler Handler) Handle(connection netconn.Context, packet codec.Packet) error {
	ctx := context.Background()
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, inRoom := player.CurrentRoom()
	switch packet.Header {
	case instatus.Header:
		if err = instatus.Decode(packet); err != nil || !inRoom {
			return err
		}
		return handler.sendStatus(ctx, connection, roomID)
	case inrent.Header:
		if err = inrent.Decode(packet); err != nil || !inRoom {
			return err
		}
		state, found, findErr := handler.Service.Status(ctx, roomID)
		if findErr != nil || !found {
			return handler.sendFailure(ctx, connection)
		}
		state, err = handler.Service.Rent(ctx, state.ItemID, player.ID())
		if err != nil {
			return handler.sendFailure(ctx, connection)
		}
		return handler.sendRentOK(ctx, connection, state)
	case incancel.Header:
		if err = incancel.Decode(packet); err != nil || !inRoom {
			return err
		}
		state, found, findErr := handler.Service.Status(ctx, roomID)
		if findErr != nil || !found {
			return handler.sendFailure(ctx, connection)
		}
		if err = handler.Service.Cancel(ctx, state.ItemID, player.ID()); err != nil {
			return handler.sendFailure(ctx, connection)
		}
		return handler.sendStatus(ctx, connection, roomID)
	case inextend.Header:
		value, decodeErr := inextend.Decode(packet)
		if decodeErr != nil {
			return decodeErr
		}
		return handler.mutateItem(ctx, connection, player.ID(), int64(value.ItemID), value.Buyout)
	case inextendstrip.Header:
		value, decodeErr := inextendstrip.Decode(packet)
		if decodeErr != nil {
			return decodeErr
		}
		return handler.mutateItem(ctx, connection, player.ID(), int64(value.ItemID), value.Buyout)
	case inoffer.Header:
		value, decodeErr := inoffer.Decode(packet)
		if decodeErr != nil {
			return decodeErr
		}
		config := handler.Service.Config()
		price := config.PriceCredits
		if value.Buyout {
			price = config.BuyoutCredits
		}
		response, encodeErr := oukoffer.Encode(value.Wall, value.ProductName, value.Buyout, price, 0, 0)
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, response)
	default:
		return codec.ErrUnexpectedHeader
	}
}

// mutateItem extends or buys out one item.
func (handler Handler) mutateItem(ctx context.Context, connection netconn.Context, playerID int64, itemID int64, buyout bool) error {
	var err error
	if buyout {
		err = handler.Service.Buyout(ctx, itemID, playerID)
	} else {
		_, err = handler.Service.Rent(ctx, itemID, playerID)
	}
	if err != nil {
		return handler.sendFailure(ctx, connection)
	}
	state, found, findErr := handler.Service.store.FindItem(ctx, itemID)
	if findErr != nil || !found {
		return findErr
	}
	return handler.sendRentOK(ctx, connection, state)
}

// sendStatus sends current room rental state.
func (handler Handler) sendStatus(ctx context.Context, connection netconn.Context, roomID int64) error {
	state, found, err := handler.Service.Status(ctx, roomID)
	if err != nil {
		return err
	}
	if !found {
		return handler.sendFailure(ctx, connection)
	}
	now := handler.Service.now()
	rented := state.ActiveAt(now)
	renterID := int32(-1)
	if state.RenterPlayerID != nil {
		renterID = int32(*state.RenterPlayerID)
	}
	response, err := outstatus.Encode(rented, 0, renterID, "", state.SecondsRemaining(now), state.PriceCredits)
	if err != nil {
		return err
	}
	return connection.Send(ctx, response)
}

// sendRentOK sends the authoritative rental expiration.
func (handler Handler) sendRentOK(ctx context.Context, connection netconn.Context, state State) error {
	expiry := int32(0)
	if state.ExpiresAt != nil {
		expiry = int32(state.ExpiresAt.Unix())
	}
	response, err := outok.Encode(expiry)
	if err != nil {
		return err
	}
	return connection.Send(ctx, response)
}

// sendFailure sends a generic guarded-rental conflict.
func (handler Handler) sendFailure(ctx context.Context, connection netconn.Context) error {
	response, err := outfailed.Encode(FailureUnavailable)
	if err != nil {
		return err
	}
	return connection.Send(ctx, response)
}
