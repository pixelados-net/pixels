package core

import (
	"context"
	"strings"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
)

// SavePreference saves navigator preferences.
func (service *Service) SavePreference(ctx context.Context, preference navmodel.Preference) (navmodel.Preference, error) {
	if err := validatePreference(preference); err != nil {
		return navmodel.Preference{}, err
	}

	return service.store.UpsertPreference(ctx, navmodel.PreferenceParams{Preference: preference})
}

// Preference finds or returns default navigator preferences.
func (service *Service) Preference(ctx context.Context, playerID int64) (navmodel.Preference, error) {
	if playerID <= 0 {
		return navmodel.Preference{}, ErrInvalidPlayerID
	}

	preference, found, err := service.store.FindPreference(ctx, playerID)
	if err != nil || found {
		return preference, err
	}

	return DefaultPreference(playerID), nil
}

// SaveCategoryPreference saves result-list display state.
func (service *Service) SaveCategoryPreference(ctx context.Context, preference navmodel.CategoryPreference) (navmodel.CategoryPreference, error) {
	preference.Code = strings.TrimSpace(preference.Code)
	if preference.PlayerID <= 0 {
		return navmodel.CategoryPreference{}, ErrInvalidPlayerID
	}
	if preference.Code == "" {
		return navmodel.CategoryPreference{}, ErrInvalidPreference
	}

	return service.store.UpsertCategoryPreference(ctx, navmodel.CategoryPreferenceParams{Preference: preference})
}

// ListCategoryPreferences lists result-list display state.
func (service *Service) ListCategoryPreferences(ctx context.Context, playerID int64) ([]navmodel.CategoryPreference, error) {
	if playerID <= 0 {
		return nil, ErrInvalidPlayerID
	}

	return service.store.ListCategoryPreferences(ctx, playerID)
}

// ListLiftedRooms lists currently active lifted rooms.
func (service *Service) ListLiftedRooms(ctx context.Context) ([]navmodel.LiftedRoom, error) {
	return service.store.ListLiftedRooms(ctx)
}

// DefaultPreference returns default navigator preferences.
func DefaultPreference(playerID int64) navmodel.Preference {
	return navmodel.Preference{
		PlayerID:        playerID,
		WindowX:         DefaultWindowX,
		WindowY:         DefaultWindowY,
		WindowWidth:     DefaultWindowWidth,
		WindowHeight:    DefaultWindowHeight,
		ResultsMode:     0,
		LeftPanelHidden: false,
	}
}

// validatePreference validates navigator preferences.
func validatePreference(preference navmodel.Preference) error {
	if preference.PlayerID <= 0 {
		return ErrInvalidPlayerID
	}
	if preference.WindowWidth <= 0 || preference.WindowHeight <= 0 {
		return ErrInvalidPreference
	}

	return nil
}
