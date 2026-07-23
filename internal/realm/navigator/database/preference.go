package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
)

const (
	// preferenceColumns contains the shared preference select list.
	preferenceColumns = `player_id, window_x, window_y, window_width, window_height, left_panel_hidden, results_mode, created_at, updated_at`

	// upsertPreferenceSQL saves navigator preferences.
	upsertPreferenceSQL = `
insert into navigator_preferences (player_id, window_x, window_y, window_width, window_height, left_panel_hidden, results_mode)
values ($1, $2, $3, $4, $5, $6, $7)
on conflict (player_id) do update set window_x = excluded.window_x, window_y = excluded.window_y, window_width = excluded.window_width, window_height = excluded.window_height, left_panel_hidden = excluded.left_panel_hidden, results_mode = excluded.results_mode, updated_at = now()
returning ` + preferenceColumns

	// findPreferenceSQL reads navigator preferences.
	findPreferenceSQL = `select ` + preferenceColumns + ` from navigator_preferences where player_id = $1`

	// categoryPreferenceColumns contains the shared category preference select list.
	categoryPreferenceColumns = `player_id, code, collapsed, list_mode, created_at, updated_at`

	// upsertCategoryPreferenceSQL saves category preferences.
	upsertCategoryPreferenceSQL = `
insert into navigator_category_preferences (player_id, code, collapsed, list_mode)
values ($1, $2, $3, $4)
on conflict (player_id, code) do update set collapsed = excluded.collapsed, list_mode = excluded.list_mode, updated_at = now()
returning ` + categoryPreferenceColumns

	// listCategoryPreferencesSQL reads category preferences.
	listCategoryPreferencesSQL = `select ` + categoryPreferenceColumns + ` from navigator_category_preferences where player_id = $1 order by code asc`
)

// UpsertPreference saves navigator preferences.
func (repository *Repository) UpsertPreference(ctx context.Context, params navmodel.PreferenceParams) (navmodel.Preference, error) {
	preference := params.Preference
	return scanPreference(repository.executor.QueryRow(ctx, upsertPreferenceSQL, preference.PlayerID, preference.WindowX, preference.WindowY, preference.WindowWidth, preference.WindowHeight, preference.LeftPanelHidden, preference.ResultsMode))
}

// FindPreference finds navigator preferences.
func (repository *Repository) FindPreference(ctx context.Context, playerID int64) (navmodel.Preference, bool, error) {
	preference, err := scanPreference(repository.executor.QueryRow(ctx, findPreferenceSQL, playerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return navmodel.Preference{}, false, nil
	}

	if err != nil {
		return navmodel.Preference{}, false, fmt.Errorf("find navigator preference: %w", err)
	}

	return preference, true, nil
}

// UpsertCategoryPreference saves result-list display state.
func (repository *Repository) UpsertCategoryPreference(ctx context.Context, params navmodel.CategoryPreferenceParams) (navmodel.CategoryPreference, error) {
	preference := params.Preference
	return scanCategoryPreference(repository.executor.QueryRow(ctx, upsertCategoryPreferenceSQL, preference.PlayerID, preference.Code, preference.Collapsed, preference.ListMode))
}

// ListCategoryPreferences lists result-list display state.
func (repository *Repository) ListCategoryPreferences(ctx context.Context, playerID int64) ([]navmodel.CategoryPreference, error) {
	rows, err := repository.executor.Query(ctx, listCategoryPreferencesSQL, playerID)
	if err != nil {
		return nil, fmt.Errorf("list navigator category preferences: %w", err)
	}
	defer rows.Close()

	return scanCategoryPreferences(rows)
}
