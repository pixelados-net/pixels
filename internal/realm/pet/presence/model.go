package presence

import (
	"context"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// PlaceParams stores one placement request.
type PlaceParams struct {
	// PetID identifies the inventory pet.
	PetID int64
	// ActorPlayerID identifies the actor.
	ActorPlayerID int64
	// RoomID identifies the current room.
	RoomID int64
	// Point stores the requested tile.
	Point grid.Point
}

// PickupParams stores one pickup request.
type PickupParams struct {
	// PetID identifies the placed pet.
	PetID int64
	// ActorPlayerID identifies the actor.
	ActorPlayerID int64
	// RoomID identifies the current room.
	RoomID int64
}

// TransitionHook appends a transaction-scoped mutation before live projection.
type TransitionHook func(context.Context, petrecord.Pet) error

// Inventory lists one owner's current inventory pets.
func (service *Service) Inventory(ctx context.Context, playerID int64) ([]petrecord.Pet, error) {
	return service.inventory.List(ctx, playerID)
}

// Find returns one live durable pet.
func (service *Service) Find(ctx context.Context, petID int64) (petrecord.Pet, bool, error) {
	return service.store.Find(ctx, petID)
}
