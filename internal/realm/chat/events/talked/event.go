// Package talked defines a committed room talk event.
package talked

import (
	"time"

	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies a delivered room talk message.
	Name bus.Name = "chat.talked"
)

// Payload describes one delivered room talk message.
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
