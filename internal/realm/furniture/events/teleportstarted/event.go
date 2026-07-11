// Package teleportstarted contains the furniture teleport start event.
package teleportstarted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the furniture teleport start event.
const Name bus.Name = "furniture.teleport_started"

// Payload describes one accepted teleport transition.
type Payload struct {
	// PlayerID identifies the moving player.
	PlayerID int64
	// SourceItemID identifies the used teleport.
	SourceItemID int64
	// SourceRoomID identifies the source room.
	SourceRoomID int64
	// TargetItemID identifies the paired teleport.
	TargetItemID int64
	// TargetRoomID identifies the destination room.
	TargetRoomID int64
}
