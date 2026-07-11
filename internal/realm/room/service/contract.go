package service

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
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

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)

// configManagerAssertion verifies Service implements ConfigManager.
var configManagerAssertion ConfigManager = (*Service)(nil)
