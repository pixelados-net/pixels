// Package group owns social group membership used by room capabilities.
package group

import "context"

// Store reads room-linked group membership snapshots.
type Store interface {
	// RoomMembership returns the linked group and current member IDs.
	RoomMembership(context.Context, int64) (int64, []int64, bool, error)
}
