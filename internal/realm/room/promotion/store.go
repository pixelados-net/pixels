package promotion

import "context"

// Store persists room promotions and owns their transaction boundary.
type Store interface {
	// WithinTransaction runs work in one shared database transaction.
	WithinTransaction(context.Context, func(context.Context) error) error
	// Upsert creates, replaces, or extends one room promotion atomically.
	Upsert(context.Context, PurchaseParams, Config) (Promotion, error)
	// FindActiveByRoom finds one unexpired room promotion.
	FindActiveByRoom(context.Context, int64) (Promotion, bool, error)
	// FindByID finds one promotion regardless of expiry.
	FindByID(context.Context, int64) (Promotion, bool, error)
	// UpdateCopy changes active promotion copy for its creator.
	UpdateCopy(context.Context, EditParams) (Promotion, bool, error)
	// ActiveRoomIDs returns promoted ids from a bounded room set.
	ActiveRoomIDs(context.Context, []int64) (map[int64]struct{}, error)
	// DeleteByRoom force-cancels one room promotion.
	DeleteByRoom(context.Context, int64) (bool, error)
}
