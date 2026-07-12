// Package requestsent contains the messenger request-sent event.
package requestsent

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the messenger request-sent event.
const Name bus.Name = "messenger.request.sent"

// Payload describes a persisted friend request.
type Payload struct {
	// FromPlayerID identifies the requester.
	FromPlayerID int64
	// ToPlayerID identifies the recipient.
	ToPlayerID int64
}
