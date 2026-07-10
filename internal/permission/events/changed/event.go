// Package changed contains the permission changed event.
package changed

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies a committed permission mutation.
	Name bus.Name = "permission.changed"
)

// Payload describes players or groups affected by a permission mutation.
type Payload struct {
	// PlayerID identifies one directly affected player when positive.
	PlayerID int64
	// GroupID identifies one affected group and its descendants when present.
	GroupID *int64
}
