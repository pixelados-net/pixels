package core

import (
	"context"
	"strconv"
	"time"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	calendarevent "github.com/niflaot/pixels/internal/realm/subscription/events/calendar"
	giftevent "github.com/niflaot/pixels/internal/realm/subscription/events/gift"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// ClaimClubGift grants one earned monthly catalog gift.
func (service *Service) ClaimClubGift(ctx context.Context, playerID int64, catalogItemID int64) (catalogservice.PurchaseResult, error) {
	var result catalogservice.PurchaseResult
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		membership, found, err := service.store.FindMembership(txCtx, playerID, true)
		if err != nil || !found || !service.activeMembership(membership) || RemainingClubGifts(membership) <= 0 || membership.StartedAt == nil {
			if err != nil {
				return err
			}
			return ErrMembershipNotFound
		}
		if err := service.validateClubGift(txCtx, playerID, catalogItemID, membership); err != nil {
			return err
		}
		period := membership.StartedAt.Add(time.Duration(membership.GiftsClaimed) * time.Duration(ClubGiftPeriodSeconds) * time.Second)
		if err := service.store.InsertGiftClaim(txCtx, playerID, period, catalogItemID); err != nil {
			return err
		}
		result, err = service.catalog.Purchase(txCtx, catalogservice.PurchaseParams{PlayerID: playerID, CatalogItemID: catalogItemID, HasClub: true, Amount: 1, Free: true})
		if err != nil {
			return err
		}
		membership.GiftsClaimed++
		return service.store.UpsertMembership(txCtx, membership)
	})
	if err == nil && service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: giftevent.Name, Payload: giftevent.Payload{PlayerID: playerID, ItemID: catalogItemID}})
	}

	return result, err
}

// activeMembership reports whether one entitlement is usable now.
func (service *Service) activeMembership(membership record.Membership) bool {
	return membership.Level > record.LevelNone && membership.ExpiresAt != nil && service.now().Before(*membership.ExpiresAt)
}

// validateClubGift verifies that one offer belongs to the gift page and meets tenure.
func (service *Service) validateClubGift(ctx context.Context, playerID int64, catalogItemID int64, membership record.Membership) error {
	pages, err := service.catalog.Pages(ctx, playerID, true)
	if err != nil {
		return err
	}
	var item catalogmodel.Item
	found := false
	for _, page := range pages {
		if page.Layout != "club_gifts" {
			continue
		}
		_, items, readErr := service.catalog.Page(ctx, page.ID, playerID, true)
		if readErr != nil {
			return readErr
		}
		for _, candidate := range items {
			if candidate.ID == catalogItemID {
				item, found = candidate, true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		return ErrOfferNotFound
	}
	required, parseErr := strconv.ParseInt(item.ExtraData, 10, 32)
	lifetimeDays := membership.LifetimeActiveSeconds / int64((24*time.Hour)/time.Second)
	if parseErr != nil || lifetimeDays < required {
		return ErrOfferNotFound
	}
	return nil
}

// TargetedOffer returns the next eligible personalized offer.
func (service *Service) TargetedOffer(ctx context.Context, playerID int64, afterID int64) (record.TargetedOffer, bool, error) {
	return service.store.FindTargetedOffer(ctx, playerID, afterID)
}

// SetTargetedState records viewed or dismissed offer state.
func (service *Service) SetTargetedState(ctx context.Context, playerID int64, offerID int64, dismissed bool) error {
	return service.store.UpdateTargetedState(ctx, playerID, offerID, dismissed)
}

// PurchaseTargetedOffer buys one personalized offer under its player limit.
func (service *Service) PurchaseTargetedOffer(ctx context.Context, playerID int64, offerID int64, quantity int32) (catalogservice.PurchaseResult, error) {
	offer, found, err := service.store.FindTargetedOfferByID(ctx, playerID, offerID)
	if err != nil || !found || offer.ID != offerID || quantity <= 0 || quantity > offer.PurchaseLimit-offer.PurchasesCount {
		if err != nil {
			return catalogservice.PurchaseResult{}, err
		}
		return catalogservice.PurchaseResult{}, ErrTargetedOfferUnavailable
	}
	var result catalogservice.PurchaseResult
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		updated, updateErr := service.store.IncrementTargetedPurchase(txCtx, playerID, offerID, quantity)
		if updateErr != nil || !updated {
			if updateErr != nil {
				return updateErr
			}
			return ErrTargetedOfferUnavailable
		}
		result, updateErr = service.catalog.Purchase(txCtx, catalogservice.PurchaseParams{PlayerID: playerID, CatalogItemID: offer.CatalogItemID, HasClub: true, Amount: quantity, OverrideCredits: &offer.PriceCredits, OverridePoints: &offer.PricePoints, OverridePointsType: &offer.PointsType})
		return updateErr
	})
	return result, err
}

// OpenCalendarDoor claims one campaign reward.
func (service *Service) OpenCalendarDoor(ctx context.Context, playerID int64, campaignName string, dayNumber int32, staff bool) (record.CampaignDay, error) {
	campaign, found, err := service.store.FindCampaign(ctx, campaignName)
	if err != nil || !found {
		return record.CampaignDay{}, ErrCalendarDoorUnavailable
	}
	currentDay := int32(service.now().Sub(campaign.StartDate) / (24 * time.Hour))
	if dayNumber < 0 || dayNumber >= campaign.DayCount || !staff && dayNumber > currentDay {
		return record.CampaignDay{}, ErrCalendarDoorUnavailable
	}
	day, found, err := service.store.FindCampaignDay(ctx, campaign.ID, dayNumber)
	if err != nil || !found {
		return record.CampaignDay{}, ErrCalendarDoorUnavailable
	}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := service.store.InsertDoorClaim(txCtx, campaign.ID, playerID, dayNumber); err != nil {
			return ErrCalendarDoorUnavailable
		}
		if day.CreditsReward > 0 {
			if _, err := service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: -1, Amount: day.CreditsReward, Reason: "calendar_door", ActorKind: currencyservice.ActorSystem}); err != nil {
				return err
			}
		}
		if day.PointsReward > 0 {
			if _, err := service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: day.PointsType, Amount: day.PointsReward, Reason: "calendar_door", ActorKind: currencyservice.ActorSystem}); err != nil {
				return err
			}
		}
		if day.ProductDefinitionID != nil {
			_, err := service.furniture.Grant(txCtx, furnitureservice.GrantParams{DefinitionID: *day.ProductDefinitionID, OwnerPlayerID: playerID, Quantity: 1, ExtraData: "0"})
			return err
		}
		return nil
	})
	if err == nil && service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: calendarevent.Name, Payload: calendarevent.Payload{PlayerID: playerID, CampaignID: campaign.ID, Day: dayNumber}})
	}

	return day, err
}

// CalendarData returns campaign, rewards, and opened doors.
func (service *Service) CalendarData(ctx context.Context, playerID int64, campaignName string) (record.Campaign, []record.CampaignDay, []int32, error) {
	campaign, found, err := service.store.FindCampaign(ctx, campaignName)
	if err != nil || !found {
		return record.Campaign{}, nil, nil, ErrCalendarDoorUnavailable
	}
	days, err := service.store.ListCampaignDays(ctx, campaign.ID)
	if err != nil {
		return record.Campaign{}, nil, nil, err
	}
	opened, err := service.store.ListOpenedDays(ctx, campaign.ID, playerID)
	return campaign, days, opened, err
}

// ActiveCalendarData returns the current campaign and opened doors when available.
func (service *Service) ActiveCalendarData(ctx context.Context, playerID int64) (record.Campaign, []int32, bool, error) {
	campaign, found, err := service.store.FindActiveCampaign(ctx, service.now())
	if err != nil || !found {
		return record.Campaign{}, nil, found, err
	}
	opened, err := service.store.ListOpenedDays(ctx, campaign.ID, playerID)
	return campaign, opened, true, err
}

// SeasonalOffer returns the catalog offer linked to one date.
func (service *Service) SeasonalOffer(ctx context.Context, date time.Time) (record.SeasonalOffer, bool, error) {
	return service.store.FindSeasonalOffer(ctx, date)
}
