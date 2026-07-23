package record

import "time"

// BreedingSession stores one nest-owned durable breeding workflow.
type BreedingSession struct {
	// NestItemID identifies the breeding nest.
	NestItemID int64
	// RoomID identifies the active room.
	RoomID int64
	// GenerationToken invalidates stale room generations.
	GenerationToken string
	// ParentOneID identifies the first parent.
	ParentOneID int64
	// ParentTwoID identifies the second parent.
	ParentTwoID int64
	// OwnerOneConfirmed reports first-owner approval.
	OwnerOneConfirmed bool
	// OwnerTwoConfirmed reports second-owner approval.
	OwnerTwoConfirmed bool
	// State stores requested, confirmed, completed, or cancelled.
	State string
	// ExpiresAt stores the absolute workflow deadline.
	ExpiresAt time.Time
	// Version stores optimistic session state.
	Version int64
}
