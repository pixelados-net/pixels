// Package repository persists room moderation state.
package repository

import (
	"context"
	"time"

	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
)

// TransactionWork performs moderation work in one shared transaction.
type TransactionWork func(context.Context) error

// Store persists current room moderation state.
type Store interface {
	// WithinTransaction runs work in one transaction.
	WithinTransaction(ctx context.Context, work TransactionWork) error
	// Mute creates or replaces a room mute.
	Mute(ctx context.Context, roomID int64, playerID int64, endsAt time.Time) error
	// Unmute ends an active room mute.
	Unmute(ctx context.Context, roomID int64, playerID int64, now time.Time) (bool, error)
	// Ban creates or replaces a room ban.
	Ban(ctx context.Context, roomID int64, playerID int64, endsAt time.Time) error
	// Unban ends an active room ban.
	Unban(ctx context.Context, roomID int64, playerID int64, now time.Time) (bool, error)
	// IsMuted reports whether a room mute is active.
	IsMuted(ctx context.Context, roomID int64, playerID int64, now time.Time) (bool, error)
	// IsBanned reports whether a room ban is active.
	IsBanned(ctx context.Context, roomID int64, playerID int64, now time.Time) (bool, error)
	// ListMutes lists active room mutes.
	ListMutes(ctx context.Context, roomID int64, now time.Time) ([]moderationmodel.Sanction, error)
	// ListBans lists active room bans.
	ListBans(ctx context.Context, roomID int64, now time.Time) ([]moderationmodel.Sanction, error)
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
