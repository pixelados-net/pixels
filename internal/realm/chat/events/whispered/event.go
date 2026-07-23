// Package whispered defines a delivered room whisper event.
package whispered

import (
	"time"

	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies a delivered room whisper.
	Name bus.Name = "chat.whispered"
)

// Payload describes one delivered room whisper.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// PlayerID identifies the speaker.
	PlayerID int64
	// TargetPlayerID identifies the recipient.
	TargetPlayerID int64
	// Message stores text visible to recipients.
	Message string
	// Censored reports whether any filter changed the text.
	Censored bool
	// CreatedAt stores the message time.
	CreatedAt time.Time
}
