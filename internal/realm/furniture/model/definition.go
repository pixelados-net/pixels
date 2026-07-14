// Package model contains persistent furniture records.
package model

import (
	"encoding/json"

	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// Kind describes a furniture definition category.
type Kind string

const (
	// KindFloor marks a floor-placed furniture definition.
	KindFloor Kind = "floor"

	// KindWall marks a wall-placed furniture definition.
	KindWall Kind = "wall"
)

// Definition contains durable furniture definition metadata.
type Definition struct {
	// Base contains shared durable record fields.
	sharedmodel.Base

	// SpriteID identifies the Nitro rendering class id.
	SpriteID int

	// Name is the stable technical identifier, such as chair_plasto.
	Name string

	// PublicName is the visible or debug display name.
	PublicName string

	// Kind describes the definition category.
	Kind Kind

	// Width stores the footprint width at rotation 0.
	Width int

	// Length stores the footprint length at rotation 0.
	Length int

	// StackHeight stores the height the definition adds when placed.
	StackHeight float64

	// AllowStack reports whether other items can stack on top.
	AllowStack bool

	// AllowWalk reports whether units can walk over the definition.
	AllowWalk bool

	// AllowSit reports whether the definition produces a sit status.
	AllowSit bool

	// AllowLay reports whether the definition produces a lay status.
	AllowLay bool

	// AllowInventoryStack reports whether inventory can group identical items.
	AllowInventoryStack bool

	// AllowTrade reports whether direct player trading is allowed.
	AllowTrade bool

	// AllowMarketplaceSale reports whether Marketplace listings are allowed.
	AllowMarketplaceSale bool

	// RedeemableCredits stores credits granted to the other party when traded.
	RedeemableCredits int32

	// EffectPool stores effect-giver candidates.
	EffectPool []int32

	// EffectMale stores the male effect-tile result.
	EffectMale *int32

	// EffectFemale stores the female effect-tile result.
	EffectFemale *int32

	// InteractionType names the server-side behavior extension point.
	InteractionType string

	// InteractionModesCount stores the number of visual states.
	InteractionModesCount int

	// Multiheight stores deferred variable-height configuration.
	Multiheight string

	// CustomParams stores deferred definition-specific parameters.
	CustomParams string

	// Metadata stores server-only structured data, such as sit/lay slots.
	Metadata json.RawMessage
}
