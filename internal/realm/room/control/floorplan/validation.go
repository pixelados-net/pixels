package floorplan

import (
	"errors"
	"strings"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// ErrorCode identifies one floor plan validation failure.
type ErrorCode string

const (
	// CodeInvalidMap reports malformed rows or unsupported tile characters.
	CodeInvalidMap ErrorCode = "invalid_map"
	// CodeTooLargeArea reports more than MaxTiles positions.
	CodeTooLargeArea ErrorCode = "too_large_area"
	// CodeTooLargeHeight reports more than MaxDimension rows.
	CodeTooLargeHeight ErrorCode = "too_large_height"
	// CodeTooLargeWidth reports a row wider than MaxDimension.
	CodeTooLargeWidth ErrorCode = "too_large_width"
	// CodeZeroHeight reports a map without usable tiles.
	CodeZeroHeight ErrorCode = "effective_height_zero"
	// CodeInvalidDoor reports an unavailable entry tile.
	CodeInvalidDoor ErrorCode = "invalid_door"
	// CodeInvalidDirection reports a door direction outside zero through seven.
	CodeInvalidDirection ErrorCode = "invalid_direction"
	// CodeInvalidWallThickness reports wall thickness outside minus two through one.
	CodeInvalidWallThickness ErrorCode = "invalid_wall_thickness"
	// CodeInvalidFloorThickness reports floor thickness outside minus two through one.
	CodeInvalidFloorThickness ErrorCode = "invalid_floor_thickness"
	// CodeInvalidWallHeight reports fixed wall height outside minus one through fifteen.
	CodeInvalidWallHeight ErrorCode = "invalid_wall_height"
	// CodeBlockedFurniture reports furniture affected by the geometry change.
	CodeBlockedFurniture ErrorCode = "blocked_furniture"
	// CodeTooManyPickups reports an unsafe automatic pickup request.
	CodeTooManyPickups ErrorCode = "too_many_pickups"
	// CodeCooldown reports a save attempted before its cooldown elapsed.
	CodeCooldown ErrorCode = "cooldown"
)

// ValidationErrors aggregates independent floor plan validation failures.
type ValidationErrors struct {
	// Codes stores unique failures in validation order.
	Codes []ErrorCode
}

// Error describes aggregated floor plan validation failures.
func (validation ValidationErrors) Error() string {
	values := make([]string, len(validation.Codes))
	for index, code := range validation.Codes {
		values[index] = string(code)
	}

	return "invalid room floor plan: " + strings.Join(values, ", ")
}

// Add appends a failure once.
func (validation *ValidationErrors) Add(code ErrorCode) {
	for _, existing := range validation.Codes {
		if existing == code {
			return
		}
	}
	validation.Codes = append(validation.Codes, code)
}

// Empty reports whether validation succeeded.
func (validation ValidationErrors) Empty() bool {
	return len(validation.Codes) == 0
}

// SaveParams contains editable custom floor plan input.
type SaveParams struct {
	// Heightmap stores the requested compact geometry.
	Heightmap string
	// DoorX stores the entry tile x coordinate.
	DoorX int
	// DoorY stores the entry tile y coordinate.
	DoorY int
	// DoorDirection stores the entry direction.
	DoorDirection int
	// WallThickness stores wall rendering thickness.
	WallThickness int
	// FloorThickness stores floor rendering thickness.
	FloorThickness int
	// WallHeight stores fixed wall height or -1 for automatic height.
	WallHeight int
	// AutoPickup reports whether affected furniture should return to inventory.
	AutoPickup bool
}

// Validated contains normalized floor plan input and parsed geometry.
type Validated struct {
	// Params stores normalized editable values.
	Params SaveParams
	// Grid stores parsed immutable geometry.
	Grid grid.Grid
}

// Validate validates floor plan input without mutating state.
func Validate(config Config, params SaveParams) (Validated, error) {
	params.Heightmap = normalizeHeightmap(params.Heightmap)
	validation := structuralErrors(config.Normalize(), params)
	roomGrid, err := grid.Parse(params.Heightmap, grid.WithDoor(params.DoorX, params.DoorY))
	if err != nil {
		if errors.Is(err, grid.ErrInvalidDoor) {
			validation.Add(CodeInvalidDoor)
		} else {
			validation.Add(CodeInvalidMap)
		}
		return Validated{}, validation
	}
	if config.RejectZeroEffectiveHeight && roomGrid.ValidCount() == 0 {
		validation.Add(CodeZeroHeight)
	}
	if !validation.Empty() {
		return Validated{}, validation
	}

	return Validated{Params: params, Grid: roomGrid}, nil
}

// structuralErrors validates limits that do not require parsed geometry.
func structuralErrors(config Config, params SaveParams) ValidationErrors {
	validation := ValidationErrors{}
	rows := strings.Split(params.Heightmap, "\r")
	if params.Heightmap == "" {
		validation.Add(CodeInvalidMap)
	}
	if len(rows) > MaxDimension {
		validation.Add(CodeTooLargeHeight)
	}
	width := -1
	area := 0
	for _, row := range rows {
		if row == "" || len(row) > MaxDimension {
			validation.Add(CodeTooLargeWidth)
		}
		if width < 0 {
			width = len(row)
		} else if len(row) != width {
			validation.Add(CodeInvalidMap)
		}
		area += len(row)
	}
	if area > MaxTiles {
		validation.Add(CodeTooLargeArea)
	}
	if params.DoorDirection < 0 || params.DoorDirection > 7 {
		validation.Add(CodeInvalidDirection)
	}
	if params.WallThickness < -2 || params.WallThickness > 1 {
		validation.Add(CodeInvalidWallThickness)
	}
	if params.FloorThickness < -2 || params.FloorThickness > 1 {
		validation.Add(CodeInvalidFloorThickness)
	}
	if params.WallHeight < -1 || params.WallHeight > 15 {
		validation.Add(CodeInvalidWallHeight)
	}
	if config.RejectZeroEffectiveHeight && strings.Trim(params.Heightmap, "xX\r") == "" {
		validation.Add(CodeZeroHeight)
	}

	return validation
}

// normalizeHeightmap converts supported line endings and invalid tile casing.
func normalizeHeightmap(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\r")
	value = strings.ReplaceAll(value, "\n", "\r")

	return strings.ReplaceAll(value, "X", "x")
}
