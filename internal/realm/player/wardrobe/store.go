package wardrobe

import "context"

// Store persists wardrobe outfits.
type Store interface {
	// Outfits returns all saved outfits ordered by slot.
	Outfits(context.Context, int64) ([]Outfit, error)
	// SaveOutfit atomically upserts one outfit slot.
	SaveOutfit(context.Context, int64, Outfit) error
	// Clothing returns all persistent clothing unlocks.
	Clothing(context.Context, int64) (ClothingSnapshot, error)
	// RedeemClothing atomically consumes one eligible inventory item and unlocks its sets.
	RedeemClothing(context.Context, int64, int64) (RedeemResult, error)
}
