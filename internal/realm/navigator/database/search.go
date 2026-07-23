package database

import (
	"context"
	"fmt"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
)

const (
	// saveSearchSQL inserts a saved search.
	saveSearchSQL = `
insert into navigator_saved_searches (player_id, code, filter, localization)
values ($1, $2, $3, $4)
returning id, player_id, code, filter, localization, created_at`

	// deleteSearchSQL deletes a saved search.
	deleteSearchSQL = `delete from navigator_saved_searches where player_id = $1 and id = $2`

	// listSavedSearchesSQL reads saved searches.
	listSavedSearchesSQL = `
select id, player_id, code, filter, localization, created_at
from navigator_saved_searches
where player_id = $1
order by created_at desc`
)

// SaveSearch saves a navigator search.
func (repository *Repository) SaveSearch(ctx context.Context, params navmodel.SaveSearchParams) (navmodel.SavedSearch, error) {
	search, err := scanSavedSearch(repository.executor.QueryRow(ctx, saveSearchSQL, params.PlayerID, params.Code, params.Filter, params.Localization))
	if err != nil {
		return navmodel.SavedSearch{}, fmt.Errorf("save navigator search: %w", err)
	}

	return search, nil
}

// DeleteSearch deletes a saved search.
func (repository *Repository) DeleteSearch(ctx context.Context, playerID int64, id int64) (bool, error) {
	tag, err := repository.executor.Exec(ctx, deleteSearchSQL, playerID, id)
	if err != nil {
		return false, fmt.Errorf("delete navigator search: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}

// ListSavedSearches lists saved searches for a player.
func (repository *Repository) ListSavedSearches(ctx context.Context, playerID int64) ([]navmodel.SavedSearch, error) {
	rows, err := repository.executor.Query(ctx, listSavedSearchesSQL, playerID)
	if err != nil {
		return nil, fmt.Errorf("list navigator searches: %w", err)
	}
	defer rows.Close()

	return scanSavedSearches(rows)
}
