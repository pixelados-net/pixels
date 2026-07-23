// Package namesaved contains the bot name-saved event.
package namesaved

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a saved bot name.
const Name bus.Name = "bot.settings.name_saved"

// Payload describes one saved bot name.
type Payload struct {
	// BotID identifies the configured bot.
	BotID int64
	// PlayerID identifies the actor.
	PlayerID int64
	// Name stores the accepted visible name.
	Name string
}
