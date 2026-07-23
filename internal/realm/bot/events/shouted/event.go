// Package shouted contains delivered bot shout events.
package shouted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies room-wide bot speech.
const Name bus.Name = "bot.shouted"

// Payload describes one delivered bot shout.
type Payload struct {
	// BotID identifies the speaker.
	BotID int64
	// RoomID identifies the room.
	RoomID int64
	// Message stores delivered text.
	Message string
}
