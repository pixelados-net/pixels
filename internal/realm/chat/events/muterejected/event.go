// Package muterejected defines rejected room chat telemetry.
package muterejected

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies chat rejected by room mute state.
	Name bus.Name = "chat.mute_rejected"
)

// Payload describes one mute rejection.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// PlayerID identifies the rejected speaker.
	PlayerID int64
	// Reason identifies direct mute or room mute-all.
	Reason string
}
