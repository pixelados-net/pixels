// Package messagesent contains the messenger private-message event.
package messagesent

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the messenger private-message event.
const Name bus.Name = "messenger.message.sent"

// Payload describes one accepted private message.
type Payload struct {
	// FromPlayerID identifies the sender.
	FromPlayerID int64
	// ToPlayerID identifies the recipient.
	ToPlayerID int64
	// Delivered reports whether the recipient was online.
	Delivered bool
}
