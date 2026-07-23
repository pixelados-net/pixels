// Package invitesent contains the messenger invite-sent event.
package invitesent

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the messenger invite-sent event.
const Name bus.Name = "messenger.invite.sent"

// Payload describes one delivered room invite batch.
type Payload struct {
	// FromPlayerID identifies the inviter.
	FromPlayerID int64
	// ToPlayerID identifies the recipient.
	ToPlayerID int64
	// RoomID identifies the invitation destination.
	RoomID int64
}
