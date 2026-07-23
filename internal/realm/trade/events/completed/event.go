// Package completed defines the trade.completed event.
package completed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a settled direct trade.
const Name bus.Name = "trade.completed"

// Payload describes one settled trade.
type Payload struct {
	// RoomID identifies the settlement room.
	RoomID int64
	// FirstPlayerID identifies the first participant.
	FirstPlayerID int64
	// SecondPlayerID identifies the second participant.
	SecondPlayerID int64
	// FirstItemIDs stores the first offer.
	FirstItemIDs []int64
	// SecondItemIDs stores the second offer.
	SecondItemIDs []int64
}
