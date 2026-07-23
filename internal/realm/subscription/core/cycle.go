package core

import (
	"context"
	"time"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	expiredevent "github.com/niflaot/pixels/internal/realm/subscription/events/expired"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/pkg/bus"
)

// RunCycle durably accrues memberships, records missed cycles, and expires tiers.
func (service *Service) RunCycle(ctx context.Context) error {
	now := service.now()
	memberships, err := service.store.ListDueMemberships(ctx, now, service.options.PaydayInterval, ClubGiftPeriodSeconds)
	if err != nil {
		return err
	}
	for _, membership := range memberships {
		if err := service.processMembership(ctx, membership.PlayerID, now); err != nil {
			return err
		}
	}
	return nil
}

// processMembership advances one locked membership to the supplied instant.
func (service *Service) processMembership(ctx context.Context, playerID int64, now time.Time) error {
	var committed record.Membership
	var previous record.Level
	var expired bool
	var paydayCreated bool
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		membership, found, err := service.store.FindMembership(txCtx, playerID, true)
		if err != nil || !found || membership.Level == record.LevelNone || membership.ExpiresAt == nil {
			return err
		}
		service.accrueMembership(&membership, now)
		paydayCreated, err = service.recordDuePaydays(txCtx, &membership, now)
		if err != nil {
			return err
		}
		if !now.Before(*membership.ExpiresAt) {
			previous, expired = membership.Level, true
			membership.Level = record.LevelNone
			if err := service.players.SetClub(txCtx, playerID, playermodel.Club{Level: playermodel.ClubLevelNone}); err != nil {
				return err
			}
		}
		committed = membership
		return service.store.UpsertMembership(txCtx, membership)
	})
	if err != nil {
		return err
	}
	if expired {
		service.projectLive(committed)
		if service.events != nil {
			_ = service.events.Publish(ctx, bus.Event{Name: expiredevent.Name, Payload: expiredevent.Payload{PlayerID: playerID, PreviousLevel: previous}})
		}
	}
	if paydayCreated && service.playerOnline(playerID) {
		return service.ClaimPaydays(ctx, playerID)
	}
	return nil
}

// accrueMembership advances durable active seconds without scheduler-time assumptions.
func (service *Service) accrueMembership(membership *record.Membership, now time.Time) {
	end := now
	if membership.ExpiresAt != nil && membership.ExpiresAt.Before(end) {
		end = *membership.ExpiresAt
	}
	start := membership.LastAccruedAt
	if start == nil {
		start = membership.StreakStartedAt
	}
	if start == nil {
		start = membership.StartedAt
	}
	if start != nil && end.After(*start) {
		seconds := int64(end.Sub(*start) / time.Second)
		membership.LifetimeActiveSeconds += seconds
		if membership.Level == record.LevelVIP {
			membership.LifetimeVIPSeconds += seconds
		}
	}
	membership.LastAccruedAt = &end
	membership.GiftsEarned = int32(membership.LifetimeActiveSeconds / ClubGiftPeriodSeconds)
}

// recordDuePaydays materializes every complete cycle independently.
func (service *Service) recordDuePaydays(ctx context.Context, membership *record.Membership, now time.Time) (bool, error) {
	if membership.LastPaydayAt == nil {
		membership.LastPaydayAt = &now
		return false, nil
	}
	through := now
	if membership.ExpiresAt != nil && membership.ExpiresAt.Before(through) {
		through = *membership.ExpiresAt
	}
	created := false
	for boundary := membership.LastPaydayAt.Add(service.options.PaydayInterval); !boundary.After(through); boundary = membership.LastPaydayAt.Add(service.options.PaydayInterval) {
		spent, err := service.catalog.CreditsSpentBetween(ctx, membership.PlayerID, *membership.LastPaydayAt, boundary)
		if err != nil {
			return false, err
		}
		streakDays := service.streakDays(*membership, boundary)
		streak, monthly := CalculatePayday(streakDays, spent, service.options.KickbackPercentage)
		_, err = service.store.InsertPayday(ctx, record.Payday{PlayerID: membership.PlayerID, OccurredAt: boundary,
			StreakDays: streakDays, CreditsSpent: spent, StreakBonus: streak, MonthlyBonus: monthly,
			TotalAwarded: streak + monthly, CurrencyType: service.options.PaydayCurrencyType})
		if err != nil {
			return false, err
		}
		membership.LastPaydayAt, created = &boundary, true
	}
	return created, nil
}

// streakDays returns complete uninterrupted membership days at one boundary.
func (service *Service) streakDays(membership record.Membership, boundary time.Time) int32 {
	if membership.StreakStartedAt == nil || boundary.Before(*membership.StreakStartedAt) {
		return 0
	}
	return int32(boundary.Sub(*membership.StreakStartedAt) / (24 * time.Hour))
}

// playerOnline reports whether an immediate committed claim can be projected.
func (service *Service) playerOnline(playerID int64) bool {
	if service.livePlayers == nil {
		return false
	}
	_, found := service.livePlayers.Find(playerID)
	return found
}
