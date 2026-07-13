// Package record defines direct-trade persistence contracts.
package record

import "time"

// Audit stores one completed trade.
type Audit struct {
	// ID identifies the audit row.
	ID int64
	// RoomID identifies the settlement room.
	RoomID int64
	// FirstPlayerID identifies the first participant.
	FirstPlayerID int64
	// SecondPlayerID identifies the second participant.
	SecondPlayerID int64
	// FirstIP stores the optional first participant address.
	FirstIP string
	// SecondIP stores the optional second participant address.
	SecondIP string
	// FirstItemIDs stores the first offer.
	FirstItemIDs []int64
	// SecondItemIDs stores the second offer.
	SecondItemIDs []int64
	// FirstRedeemableCredits stores value delivered to the second participant.
	FirstRedeemableCredits int64
	// SecondRedeemableCredits stores value delivered to the first participant.
	SecondRedeemableCredits int64
	// CreatedAt stores settlement time.
	CreatedAt time.Time
}
