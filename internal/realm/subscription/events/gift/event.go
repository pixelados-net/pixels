// Package gift contains the club gift claimed event.
package gift

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies one claimed monthly club gift.
	Name bus.Name = "subscription.club_gift.claimed"
)

// Payload contains one club gift claim.
type Payload struct {
	// PlayerID identifies the member.
	PlayerID int64
	// ItemID identifies the catalog reward.
	ItemID int64
}
