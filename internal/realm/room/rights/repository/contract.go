// Package repository persists room rights state.
package repository

import (
	"context"

	rightsmodel "github.com/niflaot/pixels/internal/realm/room/rights/model"
)

// TransactionWork performs room rights work in one shared transaction.
type TransactionWork func(context.Context) error

// Store persists room rights membership.
type Store interface {
	// WithinTransaction runs work in one transaction.
	WithinTransaction(ctx context.Context, work TransactionWork) error
	// Grant creates rights when absent.
	Grant(ctx context.Context, roomID int64, playerID int64, actorID int64) (bool, error)
	// Revoke removes one rights holder.
	Revoke(ctx context.Context, roomID int64, playerID int64) (bool, error)
	// RevokeAll removes and returns every rights holder.
	RevokeAll(ctx context.Context, roomID int64) ([]rightsmodel.Right, error)
	// List lists current rights holders.
	List(ctx context.Context, roomID int64) ([]rightsmodel.Right, error)
	// Exists reports whether a player holds rights.
	Exists(ctx context.Context, roomID int64, playerID int64) (bool, error)
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
