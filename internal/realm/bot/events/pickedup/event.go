// Package pickedup contains the bot picked-up event.
package pickedup

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the bot picked-up event.
const Name bus.Name = "bot.picked_up"

// Payload describes one completed pickup.
type Payload struct {
	// BotID identifies the bot.
	BotID int64
	// RoomID identifies the source room.
	RoomID int64
	// PlayerID identifies the actor.
	PlayerID int64
}
