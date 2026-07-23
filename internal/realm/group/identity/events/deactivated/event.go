// Package deactivated contains the social-group deactivated event.
package deactivated

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed social-group deactivation.
const Name bus.Name = "group.deactivated"

// Payload identifies the retained group generation.
type Payload struct {
	// GroupID identifies the deactivated group.
	GroupID int64
	// Version stores the committed retained version.
	Version int64
}
