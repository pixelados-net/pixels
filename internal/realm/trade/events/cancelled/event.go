// Package cancelled defines the trade.cancelled event.
package cancelled

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a cancelled direct trade.
const Name bus.Name = "trade.cancelled"

// Payload describes one cancelled trade.
type Payload struct {
	// RoomID identifies the trade room.
	RoomID int64
	// FirstPlayerID identifies the first participant.
	FirstPlayerID int64
	// SecondPlayerID identifies the second participant.
	SecondPlayerID int64
	// Reason stores zero for user/lifecycle cancellation or one for settlement failure.
	Reason int32
}
