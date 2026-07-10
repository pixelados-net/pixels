package moderation

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
)

// RoomFinder reads room moderation policy and ownership.
type RoomFinder interface {
	// FindByID finds a room by id.
	FindByID(ctx context.Context, roomID int64) (roommodel.Room, bool, error)
}

// RightsChecker resolves room-scoped build rights.
type RightsChecker interface {
	// HasRights reports whether a player holds room rights.
	HasRights(ctx context.Context, roomID int64, playerID int64) (bool, error)
}

// Reader reads active room moderation state.
type Reader interface {
	// IsBanned reports whether a room ban is active.
	IsBanned(ctx context.Context, roomID int64, playerID int64) (bool, error)
	// IsMuted reports whether a room mute is active.
	IsMuted(ctx context.Context, roomID int64, playerID int64) (bool, error)
	// ListBans lists active room bans.
	ListBans(ctx context.Context, roomID int64) ([]moderationmodel.Sanction, error)
	// ListMutes lists active room mutes.
	ListMutes(ctx context.Context, roomID int64) ([]moderationmodel.Sanction, error)
}

// Manager reads and mutates room moderation state.
type Manager interface {
	Reader
	// Kick immediately removes a target through event projection.
	Kick(ctx context.Context, roomID int64, actorID int64, targetID int64) error
	// Mute creates or replaces a room mute.
	Mute(ctx context.Context, roomID int64, actorID int64, targetID int64, minutes int32) error
	// Unmute ends an active room mute.
	Unmute(ctx context.Context, roomID int64, actorID int64, targetID int64) error
	// Ban creates or replaces a room ban.
	Ban(ctx context.Context, roomID int64, actorID int64, targetID int64, duration moderationmodel.BanDuration) error
	// Unban ends an active room ban.
	Unban(ctx context.Context, roomID int64, actorID int64, targetID int64) error
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
