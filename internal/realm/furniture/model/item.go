package model

import (
	"encoding/json"

	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// Item contains durable furniture instance placement and ownership state.
type Item struct {
	// Base contains shared durable record fields.
	sharedmodel.Base

	// DefinitionID references the furniture definition.
	DefinitionID int64

	// OwnerPlayerID identifies the current owner.
	OwnerPlayerID int64

	// RoomID identifies the room the item is placed in, nil when in inventory.
	RoomID *int64

	// X stores the floor tile x coordinate, nil when in inventory.
	X *int

	// Y stores the floor tile y coordinate, nil when in inventory.
	Y *int

	// Z stores the resolved placement height, nil when in inventory.
	Z *float64

	// Rotation stores the floor instance rotation.
	Rotation Rotation

	// WallPosition stores deferred wall placement state.
	WallPosition *string

	// ExtraData stores simple protocol-facing visual state.
	ExtraData string

	// Metadata stores server-only structured data, such as seed origin.
	Metadata json.RawMessage
}

// InRoom reports whether the item is currently placed in a room.
func (item Item) InRoom() bool {
	return item.RoomID != nil
}

// InInventory reports whether the item is currently unplaced inventory.
func (item Item) InInventory() bool {
	return item.RoomID == nil
}
