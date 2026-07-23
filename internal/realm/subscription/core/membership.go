package core

import (
	"context"
	"math"
	"time"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	createdevent "github.com/niflaot/pixels/internal/realm/subscription/events/created"
	expiredevent "github.com/niflaot/pixels/internal/realm/subscription/events/expired"
	extendedevent "github.com/niflaot/pixels/internal/realm/subscription/events/extended"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// Membership returns one membership.
func (service *Service) Membership(ctx context.Context, playerID int64) (record.Membership, bool, error) {
	membership, found, err := service.store.FindMembership(ctx, playerID, false)
	if err == nil && found && service.activeMembership(membership) {
		service.accrueMembership(&membership, service.now())
	}
	return membership, found, err
}

// Offers lists regular or extension-only club offers.
func (service *Service) Offers(ctx context.Context, deals bool) ([]record.Offer, error) {
	return service.store.ListOffers(ctx, deals)
}

// Subscribe creates or extends one player membership.
func (service *Service) Subscribe(ctx context.Context, playerID int64, level record.Level, duration time.Duration) (record.Membership, error) {
	if playerID <= 0 || level <= record.LevelNone || level > record.LevelVIP || duration <= 0 {
		return record.Membership{}, ErrOfferNotFound
	}
	var result record.Membership
	created := false
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		result, created, err = service.applySubscription(txCtx, playerID, level, duration)
		return err
	})
	service.publishSubscription(ctx, result, created, err)

	return result, err
}

// PurchaseOffer charges and applies one club offer.
func (service *Service) PurchaseOffer(ctx context.Context, buyerID int64, recipientID int64, offerID int64) (record.Membership, error) {
	return service.PurchaseOfferAmount(ctx, buyerID, recipientID, offerID, 1)
}

// PurchaseOfferAmount charges and applies one or more units of a club offer.
func (service *Service) PurchaseOfferAmount(ctx context.Context, buyerID int64, recipientID int64, offerID int64, amount int32) (record.Membership, error) {
	if amount <= 0 || amount > 100 {
		return record.Membership{}, ErrInvalidAmount
	}
	offer, found, err := service.store.FindOffer(ctx, offerID)
	if err != nil || !found {
		if err != nil {
			return record.Membership{}, err
		}
		return record.Membership{}, ErrOfferNotFound
	}
	maxDurationDays := int64(math.MaxInt64 / int64(24*time.Hour))
	if offer.DayCount <= 0 || int64(offer.DayCount) > maxDurationDays/int64(amount) {
		return record.Membership{}, ErrInvalidAmount
	}
	if recipientID == 0 {
		recipientID = buyerID
	}
	var result record.Membership
	created := false
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		credits, points, multiplyErr := purchasePrices(offer, amount)
		if multiplyErr != nil {
			return multiplyErr
		}
		if credits > 0 {
			if _, chargeErr := service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: buyerID, CurrencyType: -1, Amount: -credits, Reason: "club_subscription", ActorKind: currencyservice.ActorPlayer}); chargeErr != nil {
				return chargeErr
			}
		}
		if points > 0 {
			if _, chargeErr := service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: buyerID, CurrencyType: offer.PointsType, Amount: -points, Reason: "club_subscription", ActorKind: currencyservice.ActorPlayer}); chargeErr != nil {
				return chargeErr
			}
		}
		level := record.LevelHC
		if offer.VIP {
			level = record.LevelVIP
		}
		var applyErr error
		days := int64(offer.DayCount) * int64(amount)
		result, created, applyErr = service.applySubscription(txCtx, recipientID, level, time.Duration(days)*24*time.Hour)
		return applyErr
	})
	service.publishSubscription(ctx, result, created, err)
	return result, err
}

// PurchaseExtensionOffer purchases one extension offer matching the requested tier.
func (service *Service) PurchaseExtensionOffer(ctx context.Context, playerID int64, offerID int64, vip bool) (record.Membership, error) {
	offer, found, err := service.store.FindOffer(ctx, offerID)
	if err != nil {
		return record.Membership{}, err
	}
	if !found || !offer.Deal || offer.VIP != vip {
		return record.Membership{}, ErrOfferNotFound
	}

	return service.PurchaseOffer(ctx, playerID, playerID, offerID)
}

// purchasePrices calculates checked subscription prices.
func purchasePrices(offer record.Offer, amount int32) (int64, int64, error) {
	multiplier := int64(amount)
	if offer.PriceCredits > math.MaxInt64/multiplier || offer.PricePoints > math.MaxInt64/multiplier {
		return 0, 0, ErrInvalidAmount
	}

	return offer.PriceCredits * multiplier, offer.PricePoints * multiplier, nil
}

// applySubscription creates or extends membership inside the active transaction.
func (service *Service) applySubscription(ctx context.Context, playerID int64, level record.Level, duration time.Duration) (record.Membership, bool, error) {
	membership, found, err := service.store.FindMembership(ctx, playerID, true)
	if err != nil {
		return record.Membership{}, false, err
	}
	now := service.now()
	active := found && membership.Level > record.LevelNone && membership.ExpiresAt != nil && now.Before(*membership.ExpiresAt)
	if active {
		service.accrueMembership(&membership, now)
	}
	anchor := now
	if active {
		anchor = *membership.ExpiresAt
	}
	created := !found || membership.StartedAt == nil
	if created {
		membership = record.Membership{PlayerID: playerID, StartedAt: &now}
	}
	if !active {
		membership.Level = level
		membership.StreakStartedAt = &now
		membership.LastAccruedAt = &now
		membership.LastPaydayAt = &now
	} else if level > membership.Level {
		membership.Level = level
	}
	expires := anchor.Add(duration)
	membership.ExpiresAt = &expires
	if membership.LastPaydayAt == nil {
		membership.LastPaydayAt = &now
	}
	if err := service.store.UpsertMembership(ctx, membership); err != nil {
		return record.Membership{}, false, err
	}
	club := playermodel.Club{Level: playermodel.ClubLevel(membership.Level), ExpiresAt: &expires}
	if err := service.players.SetClub(ctx, playerID, club); err != nil {
		return record.Membership{}, false, err
	}

	return membership, created, nil
}

// publishSubscription emits a committed created or extended event.
func (service *Service) publishSubscription(ctx context.Context, membership record.Membership, created bool, err error) {
	if err != nil || service.events == nil || membership.ExpiresAt == nil {
		if err == nil {
			service.projectLive(membership)
		}
		return
	}
	service.projectLive(membership)
	name := extendedevent.Name
	payload := any(extendedevent.Payload{PlayerID: membership.PlayerID, Level: membership.Level, ExpiresAt: *membership.ExpiresAt})
	if created {
		name = createdevent.Name
		payload = createdevent.Payload{PlayerID: membership.PlayerID, Level: membership.Level, ExpiresAt: *membership.ExpiresAt}
	}
	_ = service.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
}

// projectLive updates one online player's committed club entitlement.
func (service *Service) projectLive(membership record.Membership) {
	if service.livePlayers == nil {
		return
	}
	player, found := service.livePlayers.Find(membership.PlayerID)
	if found {
		player.SetClub(playermodel.Club{Level: playermodel.ClubLevel(membership.Level), ExpiresAt: membership.ExpiresAt})
	}
}

// PurchaseDefaultOffer purchases the first enabled offer for one club tier.
func (service *Service) PurchaseDefaultOffer(ctx context.Context, playerID int64, vip bool) (record.Membership, error) {
	offers, err := service.Offers(ctx, true)
	if err != nil {
		return record.Membership{}, err
	}
	for _, offer := range offers {
		if offer.VIP == vip {
			return service.PurchaseOffer(ctx, playerID, playerID, offer.ID)
		}
	}
	offers, err = service.Offers(ctx, false)
	if err != nil {
		return record.Membership{}, err
	}
	for _, offer := range offers {
		if offer.VIP == vip {
			return service.PurchaseOffer(ctx, playerID, playerID, offer.ID)
		}
	}

	return record.Membership{}, ErrOfferNotFound
}

// Revoke removes one player's active club entitlement.
func (service *Service) Revoke(ctx context.Context, playerID int64) error {
	var membership record.Membership
	previous := record.LevelNone
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var found bool
		var err error
		membership, found, err = service.store.FindMembership(txCtx, playerID, true)
		if err != nil || !found {
			return err
		}
		previous = membership.Level
		service.accrueMembership(&membership, service.now())
		membership.Level, membership.ExpiresAt = record.LevelNone, nil
		membership.StreakStartedAt = nil
		if err := service.store.UpsertMembership(txCtx, membership); err != nil {
			return err
		}
		return service.players.SetClub(txCtx, playerID, playermodel.Club{Level: playermodel.ClubLevelNone})
	})
	if err != nil {
		return err
	}
	service.projectLive(membership)
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: expiredevent.Name, Payload: expiredevent.Payload{PlayerID: playerID, PreviousLevel: previous}})
	}

	return nil
}
