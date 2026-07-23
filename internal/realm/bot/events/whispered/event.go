// Package whispered contains delivered bot whisper events.
package whispered

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies private bot speech.
const Name bus.Name = "bot.whispered"

// Payload describes one delivered bot whisper.
type Payload struct {
	// BotID identifies the speaker.
	BotID int64
	// RoomID identifies the room.
	RoomID int64
	// Message stores delivered text.
	Message string
	// TargetPlayerID identifies the recipient.
	TargetPlayerID int64
}
