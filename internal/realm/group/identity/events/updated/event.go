// Package updated contains the social-group updated event.
package updated

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed social-group metadata update.
const Name bus.Name = "group.updated"

// Payload identifies the committed group generation.
type Payload struct {
	// GroupID identifies the updated group.
	GroupID int64
	// Version stores the committed group version.
	Version int64
	// Action identifies the bounded mutation kind.
	Action string
}
