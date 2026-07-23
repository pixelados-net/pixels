// Package gift contains catalog gift wrapping configuration.
package gift

import (
	_ "embed"
	"encoding/json"
)

// Options contains immutable gift wrapping choices.
type Options struct {
	// Price stores the credits surcharge for special wrapping.
	Price int32 `json:"price"`
	// Wrappers stores selectable wrapping furniture sprite identifiers.
	Wrappers []int32 `json:"wrappers"`
	// Boxes stores selectable box color identifiers.
	Boxes []int32 `json:"boxes"`
	// Ribbons stores selectable ribbon color identifiers.
	Ribbons []int32 `json:"ribbons"`
	// DefaultGifts stores free default gift furniture sprite identifiers.
	DefaultGifts []int32 `json:"defaultGifts"`
}

// embeddedOptions stores the default wrapping configuration.
//
//go:embed wrap.json
var embeddedOptions []byte

// NewOptions loads the repository-owned wrapping defaults.
func NewOptions() Options {
	options := Options{Price: 2, Wrappers: []int32{3372}, Boxes: []int32{0}, Ribbons: []int32{0}, DefaultGifts: []int32{187}}
	_ = json.Unmarshal(embeddedOptions, &options)

	return options
}

// Resolve validates client wrapping selections and returns their protocol values.
func (options Options) Resolve(spriteID int32, boxIndex int32, ribbonIndex int32) (int32, int32, bool) {
	defaultGift := contains(options.DefaultGifts, spriteID)
	if !contains(options.Wrappers, spriteID) && !defaultGift {
		return 0, 0, false
	}
	if boxIndex < 0 || ribbonIndex < 0 || int(ribbonIndex) >= len(options.Ribbons) {
		return 0, 0, false
	}
	if int(boxIndex) < len(options.Boxes) {
		return options.Boxes[boxIndex], options.Ribbons[ribbonIndex], true
	}
	if defaultGift && int(boxIndex) == len(options.Boxes) {
		return boxIndex, options.Ribbons[ribbonIndex], true
	}

	return 0, 0, false
}

// contains reports whether one wrapping list contains a sprite identifier.
func contains(values []int32, target int32) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}
