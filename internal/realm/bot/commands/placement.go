package commands

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	botlifecycle "github.com/niflaot/pixels/internal/realm/bot/lifecycle"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outboterror "github.com/niflaot/pixels/networking/outbound/bot/error"
)

const (
	// errorForbiddenHotel is Nitro's globally forbidden bot error.
	errorForbiddenHotel int32 = iota
	// errorForbiddenRoom is Nitro's room-forbidden bot error.
	errorForbiddenRoom
	// errorRoomLimit is Nitro's maximum room bots error.
	errorRoomLimit
	// errorTileOccupied is Nitro's selected tile error.
	errorTileOccupied
	// errorNameRejected is Nitro's rejected bot name error.
	errorNameRejected
)

const (
	// PlaceName identifies bot placement.
	PlaceName command.Name = "bot.place"
	// PickupName identifies bot pickup.
	PickupName command.Name = "bot.pickup"
)

// PlaceCommand places one inventory bot.
type PlaceCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// BotID identifies the inventory bot.
	BotID int64
	// X stores the requested tile coordinate.
	X int
	// Y stores the requested tile coordinate.
	Y int
}

// CommandName returns the stable command name.
func (PlaceCommand) CommandName() command.Name { return PlaceName }

// PickupCommand picks up one placed bot.
type PickupCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// BotID identifies the placed bot.
	BotID int64
}

// CommandName returns the stable command name.
func (PickupCommand) CommandName() command.Name { return PickupName }

// PlacementHandler handles bot placement and pickup.
type PlacementHandler struct {
	// Service coordinates bot behavior.
	Service *botlifecycle.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connections.
	Bindings *binding.Registry
}

// Handle places one bot in the actor's current room.
func (handler PlacementHandler) Handle(ctx context.Context, envelope command.Envelope[PlaceCommand]) error {
	resolved, err := player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := resolved.CurrentRoom()
	point, valid := grid.NewPoint(envelope.Command.X, envelope.Command.Y)
	if !found || !valid {
		return handler.softError(ctx, envelope.Command.Handler, botrecord.ErrTileNotFree)
	}
	_, err = handler.Service.Place(ctx, botlifecycle.PlaceParams{BotID: envelope.Command.BotID, ActorPlayerID: resolved.ID(), RoomID: roomID, Point: point})
	if isExpected(err) {
		return handler.softError(ctx, envelope.Command.Handler, err)
	}
	return err
}

// HandlePickup picks one bot up from the actor's current room.
func (handler PlacementHandler) HandlePickup(ctx context.Context, envelope command.Envelope[PickupCommand]) error {
	resolved, err := player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := resolved.CurrentRoom()
	if !found {
		return nil
	}
	_, err = handler.Service.Pickup(ctx, botlifecycle.PickupParams{BotID: envelope.Command.BotID, ActorPlayerID: resolved.ID(), RoomID: roomID})
	if isExpected(err) {
		return handler.softError(ctx, envelope.Command.Handler, err)
	}
	return err
}

// softError sends Nitro's native localized bot error response.
func (handler PlacementHandler) softError(ctx context.Context, connection netconn.Context, err error) error {
	packet, encodeErr := outboterror.Encode(errorCode(err))
	if encodeErr != nil {
		return encodeErr
	}
	return connection.Send(ctx, packet)
}

// isExpected reports a client-correctable bot error.
func isExpected(err error) bool {
	return errors.Is(err, botrecord.ErrBotNotFound) || errors.Is(err, botrecord.ErrNoRights) || errors.Is(err, botrecord.ErrRoomLimit) || errors.Is(err, botrecord.ErrInventoryLimit) || errors.Is(err, botrecord.ErrTileNotFree) || errors.Is(err, botrecord.ErrInvalidSkill) || errors.Is(err, botrecord.ErrConflict)
}

// errorCode maps domain failures to Nitro bot error identifiers.
func errorCode(err error) int32 {
	switch {
	case errors.Is(err, botrecord.ErrRoomLimit):
		return errorRoomLimit
	case errors.Is(err, botrecord.ErrTileNotFree):
		return errorTileOccupied
	case errors.Is(err, botrecord.ErrInvalidSkill):
		return errorNameRejected
	case errors.Is(err, botrecord.ErrNoRights):
		return errorForbiddenRoom
	default:
		return errorForbiddenHotel
	}
}
