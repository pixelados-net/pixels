package record

import "time"

// Visit contains one coalesced admitted room visit.
type Visit struct {
	// PlayerID identifies the visitor.
	PlayerID int64
	// RoomID identifies the admitted room.
	RoomID int64
	// VisitedAt stores the latest accepted entry time.
	VisitedAt time.Time
	// Increment reports whether this entry advances visit frequency.
	Increment bool
}
