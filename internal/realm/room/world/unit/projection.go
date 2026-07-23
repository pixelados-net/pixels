package unit

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

// SetRenderOffset replaces the protocol-only vertical offset from the physical tile.
func (unit *Unit) SetRenderOffset(offset grid.Height) {
	unit.renderOffset = offset
}

// RenderOffset returns the protocol-only vertical offset from the physical tile.
func (unit *Unit) RenderOffset() grid.Height {
	return unit.renderOffset
}
