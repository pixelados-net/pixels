package layout

import (
	"strings"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// Layout describes a room model available to Nitro.
type Layout struct {
	// Base contains shared durable record fields.
	sharedmodel.Base

	// Name stores the protocol model name such as model_a.
	Name string

	// TileSize stores the number of walkable/renderable tiles.
	TileSize int

	// Heightmap stores the model heightmap text.
	Heightmap string

	// DoorX stores the door tile x coordinate.
	DoorX int

	// DoorY stores the door tile y coordinate.
	DoorY int

	// DoorZ stores the door tile height.
	DoorZ int

	// DoorDirection stores the door rotation.
	DoorDirection int

	// ClubLevel stores the minimum club level required by the client UI.
	ClubLevel int

	// Enabled reports whether the layout can be used for new rooms.
	Enabled bool
}

// NormalizeName returns the protocol model name accepted by room creation.
func NormalizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if strings.HasPrefix(name, "model_") {
		return name
	}

	return "model_" + name
}

// Valid reports whether the layout can be registered.
func (layout Layout) Valid() bool {
	return NormalizeName(layout.Name) == layout.Name &&
		layout.TileSize > 0 &&
		layout.Heightmap != "" &&
		layout.DoorX >= 0 &&
		layout.DoorY >= 0 &&
		layout.DoorDirection >= 0
}

// Grid parses the layout heightmap into a compact room grid.
func (layout Layout) Grid() (grid.Grid, error) {
	return grid.Parse(layout.Heightmap, grid.WithDoor(layout.DoorX, layout.DoorY))
}
