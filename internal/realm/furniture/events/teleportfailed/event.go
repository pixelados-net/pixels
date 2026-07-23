// Package teleportfailed contains the furniture teleport failure event.
package teleportfailed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the furniture teleport failure event.
const Name bus.Name = "furniture.teleport_failed"

// Payload describes one aborted teleport transition.
type Payload struct {
	// PlayerID identifies the affected player.
	PlayerID int64
	// SourceItemID identifies the used teleport.
	SourceItemID int64
	// RoomID identifies the room where the transition failed.
	RoomID int64
	// Reason stores a stable diagnostic reason.
	Reason string
}
