// Package chatsaved contains the bot chat-saved event.
package chatsaved

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies saved bot automatic chat.
const Name bus.Name = "bot.settings.chat_saved"

// Payload describes one saved bot chat configuration.
type Payload struct {
	// BotID identifies the configured bot.
	BotID int64
	// PlayerID identifies the actor.
	PlayerID int64
	// LineCount stores the accepted line count.
	LineCount int
}
