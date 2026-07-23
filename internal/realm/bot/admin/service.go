// Package admin owns protected bot support and bartender configuration.
package admin

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"

	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botlifecycle "github.com/niflaot/pixels/internal/realm/bot/lifecycle"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

const maxServeKeywordLength = 32

// Service coordinates protected bot administration.
type Service struct {
	// store persists bartender mappings.
	store botrecord.Store
	// runtime owns the in-memory serving cache.
	runtime *botcore.Service
	// lifecycle owns support reads and force pickup.
	lifecycle *botlifecycle.Service
}

// New creates protected bot administration.
func New(store botrecord.Store, runtime *botcore.Service, lifecycle *botlifecycle.Service) *Service {
	return &Service{store: store, runtime: runtime, lifecycle: lifecycle}
}

// Find returns complete durable bot support state.
func (service *Service) Find(ctx context.Context, botID int64) (botrecord.Bot, bool, error) {
	return service.lifecycle.Find(ctx, botID)
}

// ForcePickup returns one placed bot to its owner.
func (service *Service) ForcePickup(ctx context.Context, botID int64) (botrecord.Bot, error) {
	return service.lifecycle.ForcePickup(ctx, botID)
}

// ListServeItems lists current bartender keyword mappings.
func (service *Service) ListServeItems(ctx context.Context) ([]botrecord.ServeItem, error) {
	return service.store.ListServeItems(ctx)
}

// CreateServeItem validates and creates a bartender keyword mapping.
func (service *Service) CreateServeItem(ctx context.Context, keyword string, definitionID int64) (botrecord.ServeItem, error) {
	keyword, err := validateKeyword(keyword, definitionID)
	if err != nil {
		return botrecord.ServeItem{}, err
	}
	item, err := service.store.CreateServeItem(ctx, keyword, definitionID)
	if err == nil {
		service.runtime.InvalidateServeItems()
	}
	return item, err
}

// UpdateServeItem validates and changes a bartender keyword mapping.
func (service *Service) UpdateServeItem(ctx context.Context, id int64, keyword string, definitionID int64) (botrecord.ServeItem, bool, error) {
	keyword, err := validateKeyword(keyword, definitionID)
	if err != nil || id <= 0 {
		return botrecord.ServeItem{}, false, firstError(err, botrecord.ErrInvalidSkill)
	}
	item, found, err := service.store.UpdateServeItem(ctx, id, keyword, definitionID)
	if err == nil && found {
		service.runtime.InvalidateServeItems()
	}
	return item, found, err
}

// DeleteServeItem removes a bartender keyword mapping.
func (service *Service) DeleteServeItem(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, botrecord.ErrInvalidSkill
	}
	deleted, err := service.store.DeleteServeItem(ctx, id)
	if err == nil && deleted {
		service.runtime.InvalidateServeItems()
	}
	return deleted, err
}

// validateKeyword normalizes one whole-word serving trigger.
func validateKeyword(keyword string, definitionID int64) (string, error) {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" || definitionID <= 0 || utf8.RuneCountInString(keyword) > maxServeKeywordLength || strings.IndexFunc(keyword, func(value rune) bool {
		return !unicode.IsLetter(value) && !unicode.IsDigit(value) && value != '_' && value != '-'
	}) >= 0 {
		return "", botrecord.ErrInvalidSkill
	}
	return keyword, nil
}

// firstError chooses infrastructure failures before domain failures.
func firstError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}
