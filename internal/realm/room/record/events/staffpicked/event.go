// Package staffpicked contains committed staff-pick grants.
package staffpicked

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a room newly selected by staff.
const Name bus.Name = "room.staff_picked"

// Payload describes the selected room and owner.
type Payload struct {
	// RoomID identifies the selected room.
	RoomID int64
	// OwnerPlayerID identifies the achievement recipient.
	OwnerPlayerID int64
}
