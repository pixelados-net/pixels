// Package relationchanged contains the messenger relation-changed event.
package relationchanged

import (
	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies the messenger relation-changed event.
const Name bus.Name = "messenger.relation.changed"

// Payload describes one unilateral relationship update.
type Payload struct {
	// PlayerID identifies the relationship owner.
	PlayerID int64
	// FriendID identifies the marked friend.
	FriendID int64
	// Relation stores the new marker.
	Relation messengermodel.Relation
}
