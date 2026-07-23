// Package serveitemrequested contains bartender request events.
package serveitemrequested

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a matched bartender request.
const Name bus.Name = "bot.serve_item.requested"

// Payload describes one matched bartender keyword.
type Payload struct {
	// BotID identifies the bartender.
	BotID int64
	// PlayerID identifies the requesting player.
	PlayerID int64
	// Keyword stores the matched whole word.
	Keyword string
	// DefinitionID identifies the served hand item.
	DefinitionID int64
}
