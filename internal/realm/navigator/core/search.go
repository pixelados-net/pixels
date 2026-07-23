package core

import (
	"context"
	"strings"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
)

// SaveSearchParams contains saved search input.
type SaveSearchParams struct {
	// PlayerID identifies the player.
	PlayerID int64

	// Code stores the navigator context or result code.
	Code string

	// Filter stores the saved search query.
	Filter string

	// Localization stores an optional localization key.
	Localization string
}

// SaveSearch saves a navigator search.
func (service *Service) SaveSearch(ctx context.Context, params SaveSearchParams) (navmodel.SavedSearch, error) {
	params = normalizeSearch(params)
	if err := validateSearch(params); err != nil {
		return navmodel.SavedSearch{}, err
	}

	return service.store.SaveSearch(ctx, navmodel.SaveSearchParams(params))
}

// DeleteSearch deletes a saved search.
func (service *Service) DeleteSearch(ctx context.Context, playerID int64, id int64) error {
	if playerID <= 0 {
		return ErrInvalidPlayerID
	}
	if id <= 0 {
		return ErrInvalidSearchID
	}

	deleted, err := service.store.DeleteSearch(ctx, playerID, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrSearchNotFound
	}

	return nil
}

// ListSavedSearches lists saved searches for a player.
func (service *Service) ListSavedSearches(ctx context.Context, playerID int64) ([]navmodel.SavedSearch, error) {
	if playerID <= 0 {
		return nil, ErrInvalidPlayerID
	}

	return service.store.ListSavedSearches(ctx, playerID)
}

// normalizeSearch normalizes saved search input.
func normalizeSearch(params SaveSearchParams) SaveSearchParams {
	params.Code = strings.TrimSpace(params.Code)
	params.Filter = strings.TrimSpace(params.Filter)
	params.Localization = strings.TrimSpace(params.Localization)

	return params
}

// validateSearch validates saved search input.
func validateSearch(params SaveSearchParams) error {
	if params.PlayerID <= 0 {
		return ErrInvalidPlayerID
	}
	if params.Code == "" || len(params.Code) > MaxSearchCodeLength {
		return ErrInvalidSearch
	}
	if len(params.Filter) > MaxSearchFilterLength {
		return ErrInvalidSearch
	}

	return nil
}
