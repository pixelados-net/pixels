package rights

import (
	"context"

	rightsmodel "github.com/niflaot/pixels/internal/realm/room/control/rights/model"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// RoomFinder reads room ownership for authorization.
type RoomFinder interface {
	// FindByID finds a room by id.
	FindByID(ctx context.Context, roomID int64) (roommodel.Room, bool, error)
}

// TransactionWork performs room rights work in one transaction.
type TransactionWork func(context.Context) error

// Store persists room rights membership.
type Store interface {
	// WithinTransaction runs work in one transaction.
	WithinTransaction(context.Context, TransactionWork) error
	// Grant creates rights when absent.
	Grant(context.Context, int64, int64, int64) (bool, error)
	// Revoke removes one rights holder.
	Revoke(context.Context, int64, int64) (bool, error)
	// RevokeAll removes and returns every rights holder.
	RevokeAll(context.Context, int64) ([]rightsmodel.Right, error)
	// List lists current rights holders.
	List(context.Context, int64) ([]rightsmodel.Right, error)
	// Exists reports whether a player holds rights.
	Exists(context.Context, int64, int64) (bool, error)
}

// Manager grants, revokes, and resolves room build rights.
type Manager interface {
	// GrantRights grants build rights.
	GrantRights(ctx context.Context, roomID int64, actorID int64, playerID int64) error
	// RevokeRights revokes one player's rights.
	RevokeRights(ctx context.Context, roomID int64, actorID int64, playerID int64) error
	// RevokeAllRights revokes every rights holder and returns the count.
	RevokeAllRights(ctx context.Context, roomID int64, actorID int64) (int, error)
	// RelinquishRights lets a player drop their own rights.
	RelinquishRights(ctx context.Context, roomID int64, playerID int64) error
	// ListRights lists current room rights holders.
	ListRights(ctx context.Context, roomID int64) ([]rightsmodel.Right, error)
	// HasRights reports whether a player holds explicit room rights.
	HasRights(ctx context.Context, roomID int64, playerID int64) (bool, error)
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
