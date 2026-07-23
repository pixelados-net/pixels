package teleport

import (
	"sync"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
)

// Phase identifies one active teleport transition step.
type Phase uint8

const (
	// PhaseResolving reserves a player while its durable pair is resolved.
	PhaseResolving Phase = iota + 1
	// PhaseApproach waits for the unit to reach the source front tile.
	PhaseApproach
	// PhaseEnter waits for the controlled step into the source teleport.
	PhaseEnter
	// PhaseCross waits for the source opening animation.
	PhaseCross
	// PhaseForward keeps the source departure visible before cross-room navigation.
	PhaseForward
	// PhaseArrival waits until a cross-room renderer can receive destination visuals.
	PhaseArrival
	// PhaseExit waits for destination opening before walking out.
	PhaseExit
	// PhaseSettle waits for the controlled exit path to finish.
	PhaseSettle
)

// Transit stores one active room-local teleport transition.
type Transit struct {
	// PlayerID identifies the moving player.
	PlayerID int64
	// Source stores the source runtime item snapshot.
	Source worldfurniture.Item
	// Target stores the target durable placement projected for runtime use.
	Target worldfurniture.Item
	// TargetRoomID identifies the target room.
	TargetRoomID int64
	// SourceRoomID identifies the source room across navigation.
	SourceRoomID int64
	// Phase stores the current transition phase.
	Phase Phase
	// Deadline stores the earliest phase advancement time.
	Deadline time.Time
	// NextItemID stores an automatically chained walk-on teleport.
	NextItemID int64
	// Handoff reports that destination state owns the pair reservation.
	Handoff bool
}

// roomState stores active transitions for one room.
type roomState struct {
	// mutex protects transitions against packet and tick concurrency.
	mutex sync.Mutex
	// transits stores transitions by player id.
	transits map[int64]Transit
}

// pendingDestination stores a cross-room destination until room entry.
type pendingDestination struct {
	// roomID identifies the expected destination room.
	roomID int64
	// transit stores the source and paired destination snapshot.
	transit Transit
	// expiresAt bounds the pending navigation handoff.
	expiresAt time.Time
}

// StartRequest contains one validated teleport start.
type StartRequest struct {
	// PlayerID identifies the moving player.
	PlayerID int64
	// Room stores the active source room.
	Room *roomlive.Room
	// ItemID identifies the used source item.
	ItemID int64
}
