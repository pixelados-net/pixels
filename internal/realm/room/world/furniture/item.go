package furniture

import (
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// Item stores runtime state for one placed furniture instance.
type Item struct {
	// ID stores the durable furniture item id, used as the surface fixture source id.
	ID int64

	// Definition stores the resolved definition snapshot for this instance.
	Definition Definition

	// Point stores the footprint origin tile at rotation 0.
	Point grid.Point

	// Z stores the base height this item sits on.
	Z grid.Height

	// Rotation stores the placed instance rotation.
	Rotation worldunit.Rotation
}

// Top returns the physical top height this item occupies once placed.
func (item Item) Top() grid.Height {
	return item.Z + item.Definition.StackHeight
}
