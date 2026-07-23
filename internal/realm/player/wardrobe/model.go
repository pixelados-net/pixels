// Package wardrobe owns persistent player avatar outfits and clothing unlocks.
package wardrobe

import "errors"

const (
	// MinSlot is the first Nitro wardrobe slot.
	MinSlot int32 = 1
	// MaxSlot is the final supported Nitro wardrobe slot.
	MaxSlot int32 = 10
)

var (
	// ErrInvalidOutfit reports a malformed slot, figure, or gender.
	ErrInvalidOutfit = errors.New("invalid wardrobe outfit")
	// ErrInvalidClothingItem reports a missing or ineligible clothing furniture.
	ErrInvalidClothingItem = errors.New("invalid clothing item")
)

// Outfit stores one persistent wardrobe slot.
type Outfit struct {
	// SlotID identifies the Nitro wardrobe slot.
	SlotID int32
	// Figure stores the avatar figure string.
	Figure string
	// Gender stores the avatar gender code.
	Gender string
}

// ClothingSnapshot contains one player's complete clothing unlock projection.
type ClothingSnapshot struct {
	// FigureSetIDs stores unlocked avatar figure set identifiers.
	FigureSetIDs []int32
	// ProductCodes stores bound clothing furniture product codes.
	ProductCodes []string
}

// RedeemResult contains one clothing redemption outcome.
type RedeemResult struct {
	// Applied reports whether the inventory item was consumed.
	Applied bool
	// Snapshot stores complete unlock state after the attempt.
	Snapshot ClothingSnapshot
}
