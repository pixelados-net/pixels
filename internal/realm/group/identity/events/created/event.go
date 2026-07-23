// Package created contains the social-group created event.
package created

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed social-group creation.
const Name bus.Name = "group.created"

// Payload identifies the committed group and owner.
type Payload struct {
	// GroupID identifies the created group.
	GroupID int64
	// OwnerPlayerID identifies the initial owner.
	OwnerPlayerID int64
	// Version stores the committed group version.
	Version int64
}
