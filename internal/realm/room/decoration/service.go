package decoration

import (
	"context"
	"regexp"
	"strings"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

var (
	// surfaceValuePattern accepts Nitro's numeric dot-delimited surface variants.
	surfaceValuePattern = regexp.MustCompile(`^[0-9]{1,4}(?:\.[0-9]{1,4}){0,2}$`)
)

// Store persists decorator furniture and room appearance atomically.
type Store interface {
	// ConsumeSurface consumes one owned inventory item and changes a room plane.
	ConsumeSurface(context.Context, int64, int64, int64, Surface, string) (bool, error)
	// PlacePostIt moves one owned post-it from inventory to a wall.
	PlacePostIt(context.Context, int64, int64, int64, string, string) (bool, error)
	// LoadDimmer returns the current room mood-light state.
	LoadDimmer(context.Context, int64) (DimmerState, bool, error)
	// SaveDimmer validates persistence ownership and replaces one preset.
	SaveDimmer(context.Context, int64, int64, Preset, bool) (DimmerState, bool, error)
	// ToggleDimmer changes the current mood-light enabled state.
	ToggleDimmer(context.Context, int64, int64) (DimmerState, bool, error)
}

// Service validates room decoration input before persistence.
type Service struct {
	// store persists room decoration state.
	store Store
}

// New creates room decoration behavior.
func New(store Store) *Service {
	return &Service{store: store}
}

// ApplySurface consumes a room-effect item and applies its catalog value.
func (service *Service) ApplySurface(ctx context.Context, itemID int64, playerID int64, roomID int64, surface Surface, value string) error {
	if !surface.Valid() {
		return ErrInvalidSurface
	}
	if !surfaceValuePattern.MatchString(value) {
		return ErrInvalidSurfaceValue
	}
	changed, err := service.store.ConsumeSurface(ctx, itemID, playerID, roomID, surface, value)
	if err != nil {
		return err
	}
	if !changed {
		return ErrDecorationUnavailable
	}

	return nil
}

// PlacePostIt places one inventory post-it at modern Nitro wall coordinates.
func (service *Service) PlacePostIt(ctx context.Context, itemID int64, playerID int64, roomID int64, wallPosition string) error {
	if !furnituremodel.ValidWallPosition(wallPosition) {
		return ErrInvalidWallPosition
	}
	changed, err := service.store.PlacePostIt(ctx, itemID, playerID, roomID, wallPosition, DefaultPostItData)
	if err != nil {
		return err
	}
	if !changed {
		return ErrDecorationUnavailable
	}

	return nil
}

// LoadDimmer returns three initialized presets when a dimmer exists.
func (service *Service) LoadDimmer(ctx context.Context, roomID int64) (DimmerState, bool, error) {
	return service.store.LoadDimmer(ctx, roomID)
}

// SaveDimmer replaces and optionally activates one preset.
func (service *Service) SaveDimmer(ctx context.Context, roomID int64, playerID int64, preset Preset, apply bool) (DimmerState, error) {
	preset.Color = strings.ToUpper(preset.Color)
	if !preset.Valid() {
		return DimmerState{}, ErrInvalidDimmerPreset
	}
	state, changed, err := service.store.SaveDimmer(ctx, roomID, playerID, preset, apply)
	if err != nil {
		return DimmerState{}, err
	}
	if !changed {
		return DimmerState{}, ErrDecorationUnavailable
	}

	return state, nil
}

// ToggleDimmer toggles the selected preset.
func (service *Service) ToggleDimmer(ctx context.Context, roomID int64, playerID int64) (DimmerState, error) {
	state, changed, err := service.store.ToggleDimmer(ctx, roomID, playerID)
	if err != nil {
		return DimmerState{}, err
	}
	if !changed {
		return DimmerState{}, ErrDecorationUnavailable
	}

	return state, nil
}

// Valid reports whether a surface is supported.
func (surface Surface) Valid() bool {
	return surface == SurfaceFloor || surface == SurfaceWallpaper || surface == SurfaceLandscape
}

// Valid reports whether a dimmer preset is protocol-safe.
func (preset Preset) Valid() bool {
	if preset.ID < 1 || preset.ID > 3 || preset.Brightness < MinimumBrightness || preset.Brightness > MaximumBrightness {
		return false
	}
	for _, allowed := range []string{"#74F5F5", "#0053F7", "#E759DE", "#EA4532", "#F2F851", "#82F349", "#000000"} {
		if preset.Color == allowed {
			return true
		}
	}

	return false
}
