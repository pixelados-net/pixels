package service

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// Creator creates room records.
type Creator interface {
	// Create creates a room and its tags.
	Create(ctx context.Context, params CreateParams) (roommodel.Room, error)
}

// Finder reads room records.
type Finder interface {
	// FindByID finds a room by id.
	FindByID(ctx context.Context, id int64) (roommodel.Room, bool, error)

	// ListByOwner lists rooms owned by a player.
	ListByOwner(ctx context.Context, ownerPlayerID int64) ([]roommodel.Room, error)

	// ListPopular lists popular rooms.
	ListPopular(ctx context.Context, limit int) ([]roommodel.Room, error)

	// ListHighestScore lists highest scoring rooms.
	ListHighestScore(ctx context.Context, limit int) ([]roommodel.Room, error)

	// Search searches public room navigator fields.
	Search(ctx context.Context, query string, limit int) ([]roommodel.Room, error)

	// ListTags lists normalized room tags.
	ListTags(ctx context.Context, roomID int64) ([]roommodel.Tag, error)
}

// Updater updates editable room settings.
type Updater interface {
	// Update applies a partial settings mutation with optimistic locking.
	Update(ctx context.Context, roomID int64, expectedVersion int64, params UpdateParams) (roommodel.Room, error)
}

// Manager creates, reads, and deletes room records.
type Manager interface {
	Creator
	Finder

	// SoftDelete soft deletes a room.
	SoftDelete(ctx context.Context, id int64) error

	// ListCategories lists room categories.
	ListCategories(ctx context.Context) ([]roommodel.Category, error)
}

// ConfigManager combines ordinary room reads with settings updates.
type ConfigManager interface {
	Manager
	Updater
}

// Store reads and writes persistent room records.
type Store interface {
	RoomReader
	RoomWriter
	CategoryReader
	TagReader
	TagWriter
}

// RoomReader reads persistent room records.
type RoomReader interface {
	// FindRoomByID finds an active room by id.
	FindRoomByID(context.Context, int64) (roommodel.Room, bool, error)
	// ListRoomsByOwner lists active rooms owned by a player.
	ListRoomsByOwner(context.Context, int64) ([]roommodel.Room, error)
	// ListPopularRooms lists active rooms by popularity.
	ListPopularRooms(context.Context, int) ([]roommodel.Room, error)
	// ListHighestScoreRooms lists active rooms by score.
	ListHighestScoreRooms(context.Context, int) ([]roommodel.Room, error)
	// SearchRooms searches active rooms by public navigator text.
	SearchRooms(context.Context, string, int) ([]roommodel.Room, error)
}

// RoomWriter writes persistent room records.
type RoomWriter interface {
	// CreateRoom creates a room record.
	CreateRoom(context.Context, CreateRecordParams) (roommodel.Room, error)
	// SoftDeleteRoom soft deletes a room record.
	SoftDeleteRoom(context.Context, int64) (bool, error)
	// UpdateRoom updates room settings and tags atomically.
	UpdateRoom(context.Context, UpdateRecordParams, []string) (roommodel.Room, bool, error)
}

// CategoryReader reads room categories.
type CategoryReader interface {
	// ListCategories lists active room categories.
	ListCategories(context.Context) ([]roommodel.Category, error)
}

// TagReader reads room tags.
type TagReader interface {
	// ListRoomTags lists tags for a room.
	ListRoomTags(context.Context, int64) ([]roommodel.Tag, error)
}

// TagWriter writes room tags.
type TagWriter interface {
	// ReplaceRoomTags replaces tags for a room.
	ReplaceRoomTags(context.Context, int64, []string) error
}

// CreateRecordParams contains room creation persistence data.
type CreateRecordParams struct {
	// OwnerPlayerID identifies the room owner.
	OwnerPlayerID int64
	// OwnerName stores an owner name snapshot.
	OwnerName string
	// Name is the visible room name.
	Name string
	// Description is the visible room description.
	Description string
	// ModelName identifies the room layout.
	ModelName string
	// DoorMode describes room entry behavior.
	DoorMode roommodel.DoorMode
	// MaxUsers stores room capacity.
	MaxUsers int
	// CategoryID optionally identifies a navigator category.
	CategoryID *int64
	// TradeMode describes room trading behavior.
	TradeMode roommodel.TradeMode
}

// UpdateRecordParams contains a complete editable room snapshot.
type UpdateRecordParams struct {
	// Room contains updated room values.
	Room roommodel.Room
	// ExpectedVersion prevents lost concurrent updates.
	ExpectedVersion int64
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)

// configManagerAssertion verifies Service implements ConfigManager.
var configManagerAssertion ConfigManager = (*Service)(nil)
