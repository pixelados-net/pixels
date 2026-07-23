// Package admin contains subscription administration behavior.
package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/niflaot/pixels/internal/realm/subscription/core"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

var (
	// ErrInvalidInput reports malformed administration input.
	ErrInvalidInput = errors.New("invalid subscription administration input")
	// ErrNotFound reports a missing administration record.
	ErrNotFound = errors.New("subscription administration record not found")
)

// Store persists subscription administration records.
type Store interface {
	// ListPaydays lists player payday history.
	ListPaydays(context.Context, int64) ([]record.Payday, error)
	// ListAllOffers lists club offers including disabled records.
	ListAllOffers(context.Context) ([]record.Offer, error)
	// UpsertOffer creates or updates a club offer.
	UpsertOffer(context.Context, record.Offer) (record.Offer, error)
	// ListTargetedOffers lists targeted offers.
	ListTargetedOffers(context.Context) ([]record.TargetedOffer, error)
	// UpsertTargetedOffer creates or updates a targeted offer.
	UpsertTargetedOffer(context.Context, record.TargetedOffer) (record.TargetedOffer, error)
	// ListCampaigns lists calendar campaigns.
	ListCampaigns(context.Context) ([]record.Campaign, error)
	// UpsertCampaign creates or updates a calendar campaign.
	UpsertCampaign(context.Context, record.Campaign) (record.Campaign, error)
	// UpsertCampaignDay creates or updates one campaign day.
	UpsertCampaignDay(context.Context, record.CampaignDay) error
}

// Service coordinates subscription administration.
type Service struct {
	// store persists administration records.
	store Store
	// subscriptions manages live membership projections.
	subscriptions *core.Service
}

// New creates subscription administration behavior.
func New(store Store, subscriptions *core.Service) *Service {
	return &Service{store: store, subscriptions: subscriptions}
}

// Membership returns membership, payday projection, and durable history.
func (service *Service) Membership(ctx context.Context, playerID int64) (record.Membership, core.PaydayInfo, []record.Payday, bool, error) {
	membership, found, err := service.subscriptions.Membership(ctx, playerID)
	if err != nil || !found {
		return membership, core.PaydayInfo{}, nil, found, err
	}
	paydays, err := service.store.ListPaydays(ctx, playerID)
	if err != nil {
		return membership, core.PaydayInfo{}, nil, true, err
	}
	projection, err := service.subscriptions.CurrentPaydayInfo(ctx, playerID)
	return membership, projection, paydays, true, err
}

// Grant grants or extends a membership.
func (service *Service) Grant(ctx context.Context, playerID int64, level record.Level, duration time.Duration) (record.Membership, error) {
	if playerID <= 0 || duration <= 0 {
		return record.Membership{}, ErrInvalidInput
	}

	return service.subscriptions.Subscribe(ctx, playerID, level, duration)
}

// Revoke revokes one membership.
func (service *Service) Revoke(ctx context.Context, playerID int64) error {
	if playerID <= 0 {
		return ErrInvalidInput
	}

	return service.subscriptions.Revoke(ctx, playerID)
}

// ClubOffers lists every club offer.
func (service *Service) ClubOffers(ctx context.Context) ([]record.Offer, error) {
	return service.store.ListAllOffers(ctx)
}

// SaveClubOffer validates and stores one club offer.
func (service *Service) SaveClubOffer(ctx context.Context, offer record.Offer) (record.Offer, error) {
	if offer.Name == "" || offer.DayCount <= 0 || offer.PriceCredits < 0 || offer.PricePoints < 0 {
		return record.Offer{}, ErrInvalidInput
	}

	return service.store.UpsertOffer(ctx, offer)
}

// TargetedOffers lists targeted offers.
func (service *Service) TargetedOffers(ctx context.Context) ([]record.TargetedOffer, error) {
	return service.store.ListTargetedOffers(ctx)
}

// SaveTargetedOffer validates and stores one targeted offer.
func (service *Service) SaveTargetedOffer(ctx context.Context, offer record.TargetedOffer) (record.TargetedOffer, error) {
	if offer.CatalogItemID <= 0 || offer.PurchaseLimit <= 0 || offer.PriceCredits < 0 || offer.PricePoints < 0 ||
		strings.TrimSpace(offer.TitleKey) == "" || strings.TrimSpace(offer.DescriptionKey) == "" ||
		strings.TrimSpace(offer.ImageURL) == "" || strings.TrimSpace(offer.IconURL) == "" ||
		offer.ExpiresAt == nil || !offer.ExpiresAt.After(time.Now()) {
		return record.TargetedOffer{}, ErrInvalidInput
	}

	return service.store.UpsertTargetedOffer(ctx, offer)
}

// Campaigns lists calendar campaigns.
func (service *Service) Campaigns(ctx context.Context) ([]record.Campaign, error) {
	return service.store.ListCampaigns(ctx)
}

// SaveCampaign validates and stores one campaign and optional days.
func (service *Service) SaveCampaign(ctx context.Context, campaign record.Campaign, days []record.CampaignDay) (record.Campaign, error) {
	if campaign.Name == "" || campaign.StartDate.IsZero() || campaign.DayCount <= 0 {
		return record.Campaign{}, ErrInvalidInput
	}
	saved, err := service.store.UpsertCampaign(ctx, campaign)
	if err != nil {
		return record.Campaign{}, err
	}
	for _, day := range days {
		day.CampaignID = saved.ID
		if day.DayNumber < 0 || day.DayNumber >= saved.DayCount {
			return record.Campaign{}, ErrInvalidInput
		}
		if err := service.store.UpsertCampaignDay(ctx, day); err != nil {
			return record.Campaign{}, err
		}
	}

	return saved, nil
}
