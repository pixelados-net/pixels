// Package shouted defines a committed room shout event.
package shouted

import (
	"time"

	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies a delivered room shout.
	Name bus.Name = "chat.shouted"
)

// Payload describes one delivered room shout.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// PlayerID identifies the speaker.
	PlayerID int64
	// Message stores text visible to recipients.
	Message string
	// Censored reports whether any filter changed the text.
	Censored bool
	// CreatedAt stores the message time.
	CreatedAt time.Time
}
