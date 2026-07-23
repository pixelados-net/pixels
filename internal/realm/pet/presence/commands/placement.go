package commands

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	petpresence "github.com/niflaot/pixels/internal/realm/pet/presence"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petsession "github.com/niflaot/pixels/internal/realm/pet/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outfailure "github.com/niflaot/pixels/networking/outbound/room/pet/place/failure"
)

const (
	// PlaceName identifies pet placement.
	PlaceName command.Name = "pet.place"
	// PickupName identifies pet pickup.
	PickupName command.Name = "pet.pickup"
)

// PlaceCommand places one inventory pet.
type PlaceCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// PetID identifies the inventory pet.
	PetID int64
	// X stores the requested tile coordinate.
	X int
	// Y stores the requested tile coordinate.
	Y int
}

// CommandName returns the stable command name.
func (PlaceCommand) CommandName() command.Name { return PlaceName }

// PickupCommand picks one room pet up.
type PickupCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// PetID identifies the room pet.
	PetID int64
}

// CommandName returns the stable command name.
func (PickupCommand) CommandName() command.Name { return PickupName }

// PlacementHandler handles pet placement and pickup.
type PlacementHandler struct {
	// Service owns placement workflows.
	Service *petpresence.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connections.
	Bindings *binding.Registry
}

// Handle places one pet in the actor's current room.
func (handler PlacementHandler) Handle(ctx context.Context, envelope command.Envelope[PlaceCommand]) error {
	player, err := petsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, inRoom := player.CurrentRoom()
	point, valid := grid.NewPoint(envelope.Command.X, envelope.Command.Y)
	if !inRoom || !valid {
		return sendPlacementFailure(ctx, envelope.Command.Handler, petrecord.ErrTileNotFree)
	}
	_, err = handler.Service.Place(ctx, petpresence.PlaceParams{PetID: envelope.Command.PetID, ActorPlayerID: player.ID(), RoomID: roomID, Point: point})
	if petpresence.IsExpected(err) {
		return sendPlacementFailure(ctx, envelope.Command.Handler, err)
	}
	return err
}

// HandlePickup returns one pet to its owner's inventory.
func (handler PlacementHandler) HandlePickup(ctx context.Context, envelope command.Envelope[PickupCommand]) error {
	player, err := petsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, inRoom := player.CurrentRoom()
	if !inRoom {
		return nil
	}
	_, err = handler.Service.Pickup(ctx, petpresence.PickupParams{PetID: envelope.Command.PetID, ActorPlayerID: player.ID(), RoomID: roomID})
	if petpresence.IsExpected(err) {
		return nil
	}
	return err
}

// sendPlacementFailure sends Nitro's native non-disconnecting placement error.
func sendPlacementFailure(ctx context.Context, target netconn.Context, err error) error {
	packet, encodeErr := outfailure.Encode(placementErrorCode(err))
	if encodeErr != nil {
		return encodeErr
	}
	return target.Send(ctx, packet)
}

// placementErrorCode maps safe domain failures to Nitro codes.
func placementErrorCode(err error) int32 {
	switch {
	case errors.Is(err, petrecord.ErrPetsDisabled):
		return 1
	case errors.Is(err, petrecord.ErrRoomLimit), errors.Is(err, petrecord.ErrInventoryLimit):
		return 2
	case errors.Is(err, petrecord.ErrTileNotFree):
		return 3
	default:
		return 0
	}
}
