// Package decoration coordinates durable room surfaces and decorator furniture.
package decoration

import "errors"

const (
	// SurfaceFloor identifies floor patterns and colors.
	SurfaceFloor Surface = "floor"
	// SurfaceWallpaper identifies wall patterns and colors.
	SurfaceWallpaper Surface = "wallpaper"
	// SurfaceLandscape identifies window views.
	SurfaceLandscape Surface = "landscape"
	// MinimumBrightness is Nitro's dimmer brightness floor.
	MinimumBrightness int32 = 76
	// MaximumBrightness is Nitro's dimmer brightness ceiling.
	MaximumBrightness int32 = 255
	// DefaultPostItColor is Nitro's classic yellow note color.
	DefaultPostItColor = "FFFF33"
	// DefaultPostItData is a complete empty note state safe to render immediately.
	DefaultPostItData = DefaultPostItColor + " "
)

var (
	// ErrInvalidSurface reports an unsupported room paint category.
	ErrInvalidSurface = errors.New("invalid room surface")
	// ErrInvalidSurfaceValue reports malformed surface catalog data.
	ErrInvalidSurfaceValue = errors.New("invalid room surface value")
	// ErrInvalidWallPosition reports malformed Nitro wall coordinates.
	ErrInvalidWallPosition = errors.New("invalid wall position")
	// ErrInvalidDimmerPreset reports unsupported preset input.
	ErrInvalidDimmerPreset = errors.New("invalid dimmer preset")
	// ErrDecorationUnavailable reports a stale or unauthorized decorator item.
	ErrDecorationUnavailable = errors.New("room decoration unavailable")
)

// Surface identifies a room plane controlled by a consumable.
type Surface string

// Appearance stores every durable room plane value.
type Appearance struct {
	// Floor stores the floor pattern and color.
	Floor string
	// Wallpaper stores the wall pattern and color.
	Wallpaper string
	// Landscape stores the window view.
	Landscape string
}

// Preset stores one mood-light preset.
type Preset struct {
	// ID identifies the one-based preset slot.
	ID int32
	// BackgroundOnly reports whether only the room background is tinted.
	BackgroundOnly bool
	// Color stores a validated uppercase hexadecimal color.
	Color string
	// Brightness stores the Nitro brightness value.
	Brightness int32
	// Selected reports whether this is the active preset.
	Selected bool
	// Enabled reports whether the mood light currently applies.
	Enabled bool
}

// DimmerState contains the updated wall item and all presets.
type DimmerState struct {
	// ItemID identifies the placed room dimmer.
	ItemID int64
	// ExtraData stores its wall-item visual state.
	ExtraData string
	// Presets stores the three ordered mood-light presets.
	Presets []Preset
}

// PostItColor returns the supported six-character sprite color from durable note data.
func PostItColor(value string) string {
	if len(value) >= 6 {
		color := value[:6]
		switch color {
		case "9CCEFF", "FF9CFF", "9CFF9C", DefaultPostItColor:
			return color
		}
	}

	return DefaultPostItColor
}
