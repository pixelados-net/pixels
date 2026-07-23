// Package promo implements promotional badge claims.
package promo

import (
	"context"
	"strings"
	"time"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionobservability "github.com/niflaot/pixels/internal/realm/progression/observability"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// BadgeGranter grants durable promotion badge rewards.
type BadgeGranter interface {
	// GrantBadge grants one badge idempotently.
	GrantBadge(context.Context, int64, string, string) (bool, error)
}

// Service owns promotional badge availability and claims.
type Service struct {
	// catalog owns immutable promotions.
	catalog *progressionengine.Catalog
	// store persists claims.
	store progressionrecord.Store
	// badges owns awarded badges.
	badges BadgeGranter
	// metrics stores process-wide progression telemetry.
	metrics *progressionobservability.Metrics
}

// SetMetrics attaches process-wide telemetry before serving claims.
func (service *Service) SetMetrics(metrics *progressionobservability.Metrics) {
	service.metrics = metrics
}

// New creates a promotional badge service.
func New(catalog *progressionengine.Catalog, store progressionrecord.Store, badges BadgeGranter) *Service {
	return &Service{catalog: catalog, store: store, badges: badges}
}

// Status reports whether one player already owns the promotion claim.
func (service *Service) Status(ctx context.Context, playerID int64, code string) (bool, error) {
	promo, found := service.find(code)
	if !found {
		return false, nil
	}
	return service.store.PromoClaimed(ctx, playerID, promo.Code)
}

// BadgeCode returns the configured badge for one enabled promotion.
func (service *Service) BadgeCode(code string) (string, bool) {
	promo, found := service.find(code)
	return promo.BadgeCode, found
}

// Claim awards one available promotion atomically.
func (service *Service) Claim(ctx context.Context, playerID int64, code string, force bool) (bool, error) {
	promo, found := service.find(code)
	if !found || !force && !available(promo, time.Now()) {
		return false, progressionrecord.ErrUnavailable
	}
	granted := false
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		granted, err = service.store.ClaimPromo(txCtx, playerID, promo, force)
		if err != nil || !granted {
			return err
		}
		_, err = service.badges.GrantBadge(txCtx, playerID, promo.BadgeCode, "promo")
		return err
	})
	if err == nil && granted {
		service.metrics.RecordReward("promo.badge")
	}
	return granted, err
}

// find resolves one enabled promotion case-insensitively.
func (service *Service) find(code string) (progressionrecord.PromoBadge, bool) {
	generation := service.catalog.Current()
	if generation == nil {
		return progressionrecord.PromoBadge{}, false
	}
	promo := generation.PromoByCode[strings.ToUpper(strings.TrimSpace(code))]
	if promo == nil {
		return progressionrecord.PromoBadge{}, false
	}
	return *promo, true
}

// available reports whether one promotion is inside its window.
func available(promo progressionrecord.PromoBadge, now time.Time) bool {
	if !promo.Enabled || promo.StartsAt != nil && now.Before(*promo.StartsAt) {
		return false
	}
	return promo.EndsAt == nil || now.Before(*promo.EndsAt)
}
