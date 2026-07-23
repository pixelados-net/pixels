// Package requestaccepted contains the messenger request-accepted event.
package requestaccepted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the messenger request-accepted event.
const Name bus.Name = "messenger.request.accepted"

// Payload describes a newly accepted friendship.
type Payload struct {
	// PlayerOneID identifies the accepting player.
	PlayerOneID int64
	// PlayerTwoID identifies the requester.
	PlayerTwoID int64
}
