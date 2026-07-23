// Package favoritechanged contains the social-group favorite changed event.
package favoritechanged

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed favorite preference change.
const Name bus.Name = "group.favorite.changed"

// Payload identifies the affected player and selected group.
type Payload struct {
	// PlayerID identifies the preference owner.
	PlayerID int64
	// GroupID identifies the selected group or zero.
	GroupID int64
}
