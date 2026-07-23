// Package placed contains the bot placed event.
package placed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the bot placed event.
const Name bus.Name = "bot.placed"

// Payload describes one completed placement.
type Payload struct {
	// BotID identifies the bot.
	BotID int64
	// RoomID identifies the destination room.
	RoomID int64
	// PlayerID identifies the actor.
	PlayerID int64
}
