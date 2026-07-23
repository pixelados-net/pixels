// Package removed contains the messenger friend-removed event.
package removed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the messenger friend-removed event.
const Name bus.Name = "messenger.friend.removed"

// Payload describes one removed friendship.
type Payload struct {
	// PlayerOneID identifies the removing player.
	PlayerOneID int64
	// PlayerTwoID identifies the removed friend.
	PlayerTwoID int64
}
