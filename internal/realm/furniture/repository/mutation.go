package repository

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// PlaceItemParams contains input for placing an owned inventory item into a room.
type PlaceItemParams struct {
	// ID identifies the furniture item.
	ID int64
	// OwnerPlayerID identifies the required current owner.
	OwnerPlayerID int64
	// RoomID identifies the destination room.
	RoomID int64
	// Placement stores the destination floor coordinates and rotation.
	Placement furnituremodel.Placement
	// WallPosition stores Nitro wall coordinates when placing a wall item.
	WallPosition string
	// UniqueInteractionType rejects duplicate room-special interactions.
	UniqueInteractionType string
}

// MoveItemParams contains input for repositioning an item within one room.
type MoveItemParams struct {
	// ID identifies the furniture item.
	ID int64
	// RoomID identifies the required current room.
	RoomID int64
	// Placement stores the destination floor coordinates and rotation.
	Placement furnituremodel.Placement
	// WallPosition stores Nitro wall coordinates when moving a wall item.
	WallPosition string
}

// PickupItemParams contains input for returning an owned, placed item to inventory.
type PickupItemParams struct {
	// ID identifies the furniture item.
	ID int64
	// OwnerPlayerID identifies the required current owner.
	OwnerPlayerID int64
}

// UpdateItemStateParams contains one guarded furniture state mutation.
type UpdateItemStateParams struct {
	// ID identifies the furniture item.
	ID int64
	// RoomID identifies the required current room.
	RoomID int64
	// Expected stores the state observed by the active room.
	Expected string
	// Next stores the state to persist.
	Next string
}

// OpenGiftItemParams contains input for opening one placed gift.
type OpenGiftItemParams struct {
	// ID identifies the furniture item.
	ID int64
	// OwnerPlayerID identifies the required current owner.
	OwnerPlayerID int64
	// RoomID identifies the required current room.
	RoomID int64
}

// PlaceItem moves an owned inventory item into a room.
func (repository *Repository) PlaceItem(ctx context.Context, params PlaceItemParams) (furnituremodel.Item, bool, error) {
	return repository.queryItem(ctx, placeItemSQL, params.ID, params.OwnerPlayerID, params.RoomID, params.Placement.X, params.Placement.Y, params.Placement.Z, params.Placement.Rotation, params.WallPosition, params.UniqueInteractionType)
}

// MoveItem repositions an item guarded by its current room.
func (repository *Repository) MoveItem(ctx context.Context, params MoveItemParams) (furnituremodel.Item, bool, error) {
	return repository.queryItem(ctx, moveItemSQL, params.ID, params.RoomID, params.Placement.X, params.Placement.Y, params.Placement.Z, params.Placement.Rotation, params.WallPosition)
}

// PickupItem returns an owned, placed item to inventory.
func (repository *Repository) PickupItem(ctx context.Context, params PickupItemParams) (furnituremodel.Item, bool, error) {
	return repository.queryItem(ctx, pickupItemSQL, params.ID, params.OwnerPlayerID)
}

// UpdateItemState changes one placed item's state with compare-and-swap semantics.
func (repository *Repository) UpdateItemState(ctx context.Context, params UpdateItemStateParams) (furnituremodel.Item, bool, error) {
	return repository.queryItem(ctx, updateItemStateSQL, params.ID, params.RoomID, params.Expected, params.Next)
}

// OpenGiftItem marks one placed gift as opened by its owner.
func (repository *Repository) OpenGiftItem(ctx context.Context, params OpenGiftItemParams) (furnituremodel.Item, bool, error) {
	return repository.queryItem(ctx, openGiftItemSQL, params.ID, params.OwnerPlayerID, params.RoomID)
}
