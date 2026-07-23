package search

import (
	"context"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// FriendIDs reads one player's directional friendship identifiers.
type FriendIDs interface {
	// FriendIDs returns one player's friend identifiers.
	FriendIDs(context.Context, int64) ([]int64, error)
}

// GroupRooms reads social-group headquarters for navigator searches.
type GroupRooms interface {
	// PlayerGroups lists active memberships.
	PlayerGroups(context.Context, int64) ([]grouprecord.PlayerGroup, error)
	// PopularGroups lists active groups by descending member count.
	PopularGroups(context.Context, int) ([]grouprecord.Group, error)
	// Group returns one active group.
	Group(context.Context, int64) (grouprecord.Group, bool, error)
}

// RightRooms reads rooms where a player holds explicit rights.
type RightRooms interface {
	// RoomIDsForPlayer lists explicit rights room identifiers.
	RoomIDsForPlayer(context.Context, int64) ([]int64, error)
}
