// Package repository contains PostgreSQL access for navigator records.
package repository

import (
	"context"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/model"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository reads and writes navigator persistence records.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor
}

// Store reads and writes navigator persistence records.
type Store interface {
	// AddFavorite adds a favorite room for a player.
	AddFavorite(ctx context.Context, playerID int64, roomID int64) error

	// RemoveFavorite removes a favorite room for a player.
	RemoveFavorite(ctx context.Context, playerID int64, roomID int64) error

	// ListFavoriteRoomIDs lists favorite room ids for a player.
	ListFavoriteRoomIDs(ctx context.Context, playerID int64) ([]int64, error)

	// SaveSearch saves a navigator search.
	SaveSearch(ctx context.Context, params SaveSearchParams) (navmodel.SavedSearch, error)

	// DeleteSearch deletes a saved search.
	DeleteSearch(ctx context.Context, playerID int64, id int64) (bool, error)

	// ListSavedSearches lists saved searches for a player.
	ListSavedSearches(ctx context.Context, playerID int64) ([]navmodel.SavedSearch, error)

	// UpsertPreference saves navigator preferences.
	UpsertPreference(ctx context.Context, params PreferenceParams) (navmodel.Preference, error)

	// FindPreference finds navigator preferences.
	FindPreference(ctx context.Context, playerID int64) (navmodel.Preference, bool, error)

	// UpsertCategoryPreference saves result-list display state.
	UpsertCategoryPreference(ctx context.Context, params CategoryPreferenceParams) (navmodel.CategoryPreference, error)

	// ListCategoryPreferences lists result-list display state.
	ListCategoryPreferences(ctx context.Context, playerID int64) ([]navmodel.CategoryPreference, error)

	// ListLiftedRooms lists currently active lifted rooms.
	ListLiftedRooms(ctx context.Context) ([]navmodel.LiftedRoom, error)
}

// New creates a navigator repository.
func New(executor postgres.Executor) *Repository {
	return &Repository{executor: executor}
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
