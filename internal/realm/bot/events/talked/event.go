// Package talked contains ordinary delivered bot speech events.
package talked

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies ordinary bot speech.
const Name bus.Name = "bot.talked"

// Payload describes one delivered ordinary bot message.
type Payload struct {
	// BotID identifies the speaker.
	BotID int64
	// RoomID identifies the room.
	RoomID int64
	// Message stores delivered text.
	Message string
}
