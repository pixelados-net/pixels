package grid

import (
	"fmt"
	"strings"
)

// Option configures heightmap parsing.
type Option func(*options)

// options stores heightmap parser settings.
type options struct {
	// door stores the optional room door coordinate.
	door Point

	// hasDoor reports whether a door coordinate was configured.
	hasDoor bool

	// validDoor reports whether the configured door coordinate is representable.
	validDoor bool
}

// WithDoor configures the room door coordinate.
func WithDoor(x int, y int) Option {
	return func(options *options) {
		point, ok := NewPoint(x, y)
		options.door = point
		options.hasDoor = true
		options.validDoor = ok
	}
}

// Parse decodes a room heightmap into a compact grid.
func Parse(heightmap string, opts ...Option) (Grid, error) {
	settings := options{}
	for _, option := range opts {
		option(&settings)
	}

	rows := normalizeRows(heightmap)
	if len(rows) == 0 {
		return Grid{}, ErrEmptyHeightmap
	}

	width := len([]rune(rows[0]))
	if width == 0 {
		return Grid{}, ErrEmptyHeightmap
	}

	roomGrid, err := parseRows(rows, width)
	if err != nil {
		return Grid{}, err
	}

	if err := applyDoor(&roomGrid, settings); err != nil {
		return Grid{}, err
	}

	return roomGrid, nil
}

// normalizeRows converts client heightmap line endings into rows.
func normalizeRows(heightmap string) []string {
	normalized := strings.TrimSpace(heightmap)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\r")
	normalized = strings.ReplaceAll(normalized, "\n", "\r")
	normalized = strings.Trim(normalized, "\r")
	if normalized == "" {
		return nil
	}

	return strings.Split(normalized, "\r")
}

// parseRows decodes normalized heightmap rows.
func parseRows(rows []string, width int) (Grid, error) {
	if width > int(^uint16(0)) || len(rows) > int(^uint16(0)) {
		return Grid{}, ErrOutOfBounds
	}

	roomGrid := Grid{
		width:   uint16(width),
		height:  uint16(len(rows)),
		heights: make([]Height, 0, width*len(rows)),
		flags:   make([]TileFlag, 0, width*len(rows)),
	}
	for rowIndex, row := range rows {
		if len([]rune(row)) != width {
			return Grid{}, ParseError{Row: rowIndex, Err: ErrIrregularRows}
		}
		if err := parseRow(&roomGrid, rowIndex, row); err != nil {
			return Grid{}, err
		}
	}

	return roomGrid, nil
}

// parseRow decodes one normalized heightmap row.
func parseRow(roomGrid *Grid, rowIndex int, row string) error {
	for columnIndex, value := range row {
		height, valid, err := parseHeight(value)
		if err != nil {
			return ParseError{Row: rowIndex, Column: columnIndex, Rune: value, Err: err}
		}

		flags := TileFlag(0)
		if !valid {
			flags = FlagInvalid
		} else {
			roomGrid.validCount++
		}
		roomGrid.heights = append(roomGrid.heights, height)
		roomGrid.flags = append(roomGrid.flags, flags)
	}

	return nil
}

// parseHeight decodes one heightmap character.
func parseHeight(value rune) (Height, bool, error) {
	if value == 'x' || value == 'X' {
		return 0, false, nil
	}
	if value >= '0' && value <= '9' {
		return HeightFromInt(int(value - '0')), true, nil
	}
	if value >= 'a' && value <= 'z' {
		return HeightFromInt(int(value-'a') + 10), true, nil
	}
	if value >= 'A' && value <= 'Z' {
		return HeightFromInt(int(value-'A') + 10), true, nil
	}

	return 0, false, fmt.Errorf("%w: %q", ErrInvalidHeight, value)
}

// applyDoor marks the configured room door.
func applyDoor(roomGrid *Grid, settings options) error {
	if !settings.hasDoor {
		return nil
	}
	if !settings.validDoor {
		return ErrInvalidDoor
	}

	index, ok := roomGrid.Index(settings.door)
	if !ok || roomGrid.flags[index]&FlagInvalid != 0 {
		return ErrInvalidDoor
	}

	roomGrid.door = settings.door
	roomGrid.hasDoor = true
	roomGrid.flags[index] |= FlagDoor

	return nil
}
