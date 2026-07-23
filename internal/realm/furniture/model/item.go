package model

import (
	"encoding/json"
	"time"

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

	// RentalOwnerPlayerID identifies the active renter while ownership remains unchanged.
	RentalOwnerPlayerID *int64

	// RentalExpiresAt stores the active rental boundary.
	RentalExpiresAt *time.Time

	// RentalPriceCredits stores the configured extension price.
	RentalPriceCredits *int32

	// StackHeightOverrideCM stores an exact stack surface height in centimeters.
	StackHeightOverrideCM *int32

	// LimitedEditionNumber stores the durable LTD serial number.
	LimitedEditionNumber *int32

	// MarketplaceReserved reports whether the item is withdrawn into Marketplace limbo.
	MarketplaceReserved bool

	// GiftWrapped reports whether this inventory item is an unopened gift.
	GiftWrapped bool

	// GiftWrapSpriteID stores the selected wrapping furniture sprite.
	GiftWrapSpriteID *int32

	// GiftWrapBoxID stores the selected wrapping box.
	GiftWrapBoxID *int32

	// GiftWrapRibbonID stores the selected ribbon.
	GiftWrapRibbonID *int32

	// GiftSenderPlayerID identifies the sender when their identity is visible.
	GiftSenderPlayerID *int64

	// GiftMessage stores the sender's message.
	GiftMessage *string

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
