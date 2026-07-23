// Package started defines the trade.started event.
package started

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a newly opened direct trade.
const Name bus.Name = "trade.started"

// Payload describes the two participants and room.
type Payload struct {
	// RoomID identifies the shared room.
	RoomID int64
	// FirstPlayerID identifies the initiator.
	FirstPlayerID int64
	// SecondPlayerID identifies the target.
	SecondPlayerID int64
}
