// Package looksaved contains the bot look-saved event.
package looksaved

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a saved bot look.
const Name bus.Name = "bot.settings.look_saved"

// Payload describes one saved bot look.
type Payload struct {
	// BotID identifies the configured bot.
	BotID int64
	// PlayerID identifies the actor.
	PlayerID int64
}
