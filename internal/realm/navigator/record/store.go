package record

import "context"

// Store persists Navigator records.
type Store interface {
	// AddFavorite adds a favorite room for a player.
	AddFavorite(ctx context.Context, playerID int64, roomID int64, limit int32, unlimited bool) error
	// RemoveFavorite removes a favorite room for a player.
	RemoveFavorite(ctx context.Context, playerID int64, roomID int64) error
	// ListFavoriteRoomIDs lists favorite room identifiers for a player.
	ListFavoriteRoomIDs(ctx context.Context, playerID int64) ([]int64, error)
	// SaveSearch saves a Navigator search.
	SaveSearch(ctx context.Context, params SaveSearchParams) (SavedSearch, error)
	// DeleteSearch deletes a saved search.
	DeleteSearch(ctx context.Context, playerID int64, id int64) (bool, error)
	// ListSavedSearches lists saved searches for a player.
	ListSavedSearches(ctx context.Context, playerID int64) ([]SavedSearch, error)
	// UpsertPreference saves Navigator preferences.
	UpsertPreference(ctx context.Context, params PreferenceParams) (Preference, error)
	// FindPreference finds Navigator preferences.
	FindPreference(ctx context.Context, playerID int64) (Preference, bool, error)
	// UpsertCategoryPreference saves result-list display state.
	UpsertCategoryPreference(ctx context.Context, params CategoryPreferenceParams) (CategoryPreference, error)
	// ListCategoryPreferences lists result-list display state.
	ListCategoryPreferences(ctx context.Context, playerID int64) ([]CategoryPreference, error)
	// ListLiftedRooms lists currently active lifted rooms.
	ListLiftedRooms(ctx context.Context) ([]LiftedRoom, error)
}

// VisitStore persists and reads bounded room visit history.
type VisitStore interface {
	// RecordVisit upserts one admitted room visit.
	RecordVisit(ctx context.Context, playerID int64, roomID int64) error
	// ListRecentRoomIDs lists room ids by latest visit.
	ListRecentRoomIDs(ctx context.Context, playerID int64, limit int) ([]int64, error)
	// ListFrequentRoomIDs lists room ids by frequency and recency.
	ListFrequentRoomIDs(ctx context.Context, playerID int64, limit int) ([]int64, error)
}

// VisitBatchStore persists coalesced room visits in one database operation.
type VisitBatchStore interface {
	// RecordVisits upserts one bounded batch of admitted visits.
	RecordVisits(ctx context.Context, visits []Visit) error
}

// VisitAdminStore removes one player's retained navigator history.
type VisitAdminStore interface {
	// DeleteVisitHistory deletes all recorded visits for a player.
	DeleteVisitHistory(context.Context, int64) (int64, error)
}

// SaveSearchParams contains saved search data.
type SaveSearchParams struct {
	// PlayerID identifies the player.
	PlayerID int64
	// Code stores the Navigator context or result code.
	Code string
	// Filter stores the saved search query.
	Filter string
	// Localization stores an optional localization key.
	Localization string
}

// PreferenceParams contains Navigator preference data.
type PreferenceParams struct {
	// Preference contains Navigator preferences.
	Preference Preference
}

// CategoryPreferenceParams contains category preference data.
type CategoryPreferenceParams struct {
	// Preference contains category preferences.
	Preference CategoryPreference
}
