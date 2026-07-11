package repository

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
)

// RoomReader reads room records.
type RoomReader interface {
	// FindRoomByID finds an active room by id.
	FindRoomByID(ctx context.Context, id int64) (roommodel.Room, bool, error)

	// ListRoomsByOwner lists active rooms owned by a player.
	ListRoomsByOwner(ctx context.Context, ownerPlayerID int64) ([]roommodel.Room, error)

	// ListPopularRooms lists active rooms by occupancy-facing popularity fields.
	ListPopularRooms(ctx context.Context, limit int) ([]roommodel.Room, error)

	// ListHighestScoreRooms lists active rooms by score.
	ListHighestScoreRooms(ctx context.Context, limit int) ([]roommodel.Room, error)

	// SearchRooms searches active rooms by public navigator text.
	SearchRooms(ctx context.Context, query string, limit int) ([]roommodel.Room, error)
}

// RoomWriter writes room records.
type RoomWriter interface {
	// CreateRoom creates a room record.
	CreateRoom(ctx context.Context, params CreateRoomParams) (roommodel.Room, error)

	// SoftDeleteRoom soft deletes a room record.
	SoftDeleteRoom(ctx context.Context, id int64) (bool, error)

	// UpdateRoom updates room settings and tags atomically with optimistic locking.
	UpdateRoom(ctx context.Context, params UpdateRoomParams, tags []string) (roommodel.Room, bool, error)
}

// CategoryReader reads room category records.
type CategoryReader interface {
	// ListCategories lists active room categories.
	ListCategories(ctx context.Context) ([]roommodel.Category, error)
}

// TagReader reads room tag records.
type TagReader interface {
	// ListRoomTags lists tags for a room.
	ListRoomTags(ctx context.Context, roomID int64) ([]roommodel.Tag, error)
}

// TagWriter writes room tag records.
type TagWriter interface {
	// ReplaceRoomTags replaces tags for a room.
	ReplaceRoomTags(ctx context.Context, roomID int64, tags []string) error
}

// Store reads and writes room persistence records.
type Store interface {
	RoomReader
	RoomWriter
	CategoryReader
	TagReader
	TagWriter
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
