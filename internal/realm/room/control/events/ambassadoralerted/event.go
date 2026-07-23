// Package ambassadoralerted defines an ambassador room-alert event.
package ambassadoralerted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a submitted ambassador room alert.
const Name bus.Name = "room.ambassador.alerted"

// Payload contains the durable moderation intake identity.
type Payload struct {
	// ReporterPlayerID identifies the ambassador.
	ReporterPlayerID int64
	// ReportedPlayerID identifies the selected room user.
	ReportedPlayerID int64
	// RoomID identifies the incident room.
	RoomID int64
}
