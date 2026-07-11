// Package settingsupdated contains the room settings updated event.
package settingsupdated

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the room settings updated event.
const Name bus.Name = "room.settings_updated"

// Payload describes a committed room settings update.
type Payload struct {
	// RoomID identifies the updated room.
	RoomID int64
	// ActorID identifies the player that saved settings.
	ActorID int64
	// Version stores the committed room version.
	Version int64
}
