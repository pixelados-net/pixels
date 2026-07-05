// Package model contains persistent room records.
package model

import sharedmodel "github.com/niflaot/pixels/pkg/model"

// DoorMode describes how a room accepts entry.
type DoorMode int16

const (
	// DoorModeOpen allows direct entry.
	DoorModeOpen DoorMode = 0

	// DoorModeDoorbell requires owner approval.
	DoorModeDoorbell DoorMode = 1

	// DoorModePassword requires a password.
	DoorModePassword DoorMode = 2

	// DoorModeInvisible hides the room from normal access.
	DoorModeInvisible DoorMode = 3
)

// TradeMode describes trading behavior inside a room.
type TradeMode int16

const (
	// TradeModeDisabled disables trading.
	TradeModeDisabled TradeMode = 0

	// TradeModeController allows controllers to decide trading.
	TradeModeController TradeMode = 1

	// TradeModeAllowed allows trading.
	TradeModeAllowed TradeMode = 2
)

// Room contains durable room metadata and settings.
type Room struct {
	// Base contains shared durable record fields.
	sharedmodel.Base

	// OwnerPlayerID identifies the player that owns the room.
	OwnerPlayerID int64

	// OwnerName stores an owner name snapshot for navigator listings.
	OwnerName string

	// Name is the visible room name.
	Name string

	// Description is the visible room description.
	Description string

	// ModelName is the room layout model name.
	ModelName string

	// DoorMode describes how the room accepts entry.
	DoorMode DoorMode

	// MaxUsers stores the maximum user count.
	MaxUsers int

	// Score stores the navigator score.
	Score int

	// CategoryID optionally identifies the navigator category.
	CategoryID *int64

	// TradeMode describes trading behavior.
	TradeMode TradeMode

	// AllowWalkthrough reports whether users can walk through each other.
	AllowWalkthrough bool

	// AllowPets reports whether pets are allowed.
	AllowPets bool

	// AllowPetsEat reports whether pets can eat.
	AllowPetsEat bool

	// HideWalls reports whether walls are hidden.
	HideWalls bool

	// WallThickness stores wall render thickness.
	WallThickness int

	// FloorThickness stores floor render thickness.
	FloorThickness int

	// ChatMode stores room chat mode.
	ChatMode int16

	// ChatWeight stores room chat weight.
	ChatWeight int16

	// ChatSpeed stores room chat speed.
	ChatSpeed int16

	// ChatDistance stores room chat distance.
	ChatDistance int16

	// ChatProtection stores room chat flood protection.
	ChatProtection int16

	// StaffPicked reports whether staff highlighted the room.
	StaffPicked bool

	// PublicRoom reports whether the room behaves as public content.
	PublicRoom bool
}
