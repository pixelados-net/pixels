// Package requestdeclined contains the messenger request-declined event.
package requestdeclined

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the messenger request-declined event.
const Name bus.Name = "messenger.request.declined"

// Payload describes a pending-request deletion.
type Payload struct {
	// FromPlayerID identifies the requester.
	FromPlayerID int64
	// ToPlayerID identifies the declining recipient.
	ToPlayerID int64
}
