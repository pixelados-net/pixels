// Package teleportcompleted contains the furniture teleport completion event.
package teleportcompleted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the furniture teleport completion event.
const Name bus.Name = "furniture.teleport_completed"

// Payload describes one completed teleport transition.
type Payload struct {
	// PlayerID identifies the moved player.
	PlayerID int64
	// SourceItemID identifies the source teleport.
	SourceItemID int64
	// SourceRoomID identifies the source room.
	SourceRoomID int64
	// TargetItemID identifies the destination teleport.
	TargetItemID int64
	// TargetRoomID identifies the destination room.
	TargetRoomID int64
}
