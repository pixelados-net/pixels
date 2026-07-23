package core

import (
	"context"
	"time"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	paydayevent "github.com/niflaot/pixels/internal/realm/subscription/events/payday"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// ClubGiftPeriodSeconds is the confirmed 31-day accumulated HC period.
	ClubGiftPeriodSeconds int64 = 2_678_400
)

// StreakBonus describes one non-cumulative reward threshold.
type StreakBonus struct {
	// Days stores the minimum streak.
	Days int32
	// Bonus stores the granted credits.
	Bonus int64
}

// PaydayInfo contains the current kickback projection.
type PaydayInfo struct {
	// Membership stores current club state.
	Membership record.Membership
	// CreditsSpent stores eligible spending since the last payday.
	CreditsSpent int64
	// StreakDays stores active calendar days.
	StreakDays int32
	// StreakBonus stores the projected streak reward.
	StreakBonus int64
	// MonthlyBonus stores the projected spending reward.
	MonthlyBonus int64
	// MinutesUntilPayday stores Nitro's next-cycle countdown unit.
	MinutesUntilPayday int32
}

var (
	// StreakBonuses stores ordered kickback thresholds.
	StreakBonuses = [...]StreakBonus{{7, 5}, {30, 10}, {60, 15}, {90, 20}, {180, 25}, {365, 30}}
)

// CalculatePayday calculates streak and spending rewards.
func CalculatePayday(streakDays int32, creditsSpent int64, percentage float64) (int64, int64) {
	streak := int64(0)
	for _, threshold := range StreakBonuses {
		if streakDays >= threshold.Days {
			streak = threshold.Bonus
		}
	}
	return streak, int64(float64(creditsSpent) * percentage)
}

// RemainingClubGifts returns unclaimed accumulated gift periods.
func RemainingClubGifts(membership record.Membership) int32 {
	earned := int32(membership.LifetimeActiveSeconds / ClubGiftPeriodSeconds)
	return earned - membership.GiftsClaimed
}

// CurrentPaydayInfo calculates one player's current kickback projection.
func (service *Service) CurrentPaydayInfo(ctx context.Context, playerID int64) (PaydayInfo, error) {
	membership, found, err := service.Membership(ctx, playerID)
	if err != nil || !found || membership.LastPaydayAt == nil {
		return PaydayInfo{}, err
	}
	now := service.now()
	spent, err := service.catalog.CreditsSpentSince(ctx, playerID, *membership.LastPaydayAt)
	if err != nil {
		return PaydayInfo{}, err
	}
	streakDays := int32(0)
	if membership.StreakStartedAt != nil {
		streakDays = int32(now.Sub(*membership.StreakStartedAt) / (24 * time.Hour))
	}
	streak, monthly := CalculatePayday(streakDays, spent, service.options.KickbackPercentage)
	remaining := service.options.PaydayInterval - now.Sub(*membership.LastPaydayAt)
	if remaining < 0 {
		remaining = 0
	}

	minutes := int32((remaining + time.Minute - 1) / time.Minute)
	return PaydayInfo{Membership: membership, CreditsSpent: spent, StreakDays: streakDays,
		StreakBonus: streak, MonthlyBonus: monthly, MinutesUntilPayday: minutes}, nil
}

// KickbackPercentage returns the configured reward fraction.
func (service *Service) KickbackPercentage() float64 { return service.options.KickbackPercentage }

// ClaimPaydays grants all pending kickback rewards exactly once.
func (service *Service) ClaimPaydays(ctx context.Context, playerID int64) error {
	var claimed []record.Payday
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		paydays, err := service.store.ListUnclaimedPaydays(txCtx, playerID)
		if err != nil {
			return err
		}
		for _, payday := range paydays {
			if payday.TotalAwarded > 0 {
				if _, grantErr := service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: payday.CurrencyType, Amount: payday.TotalAwarded, Reason: "hc_payday", ActorKind: currencyservice.ActorSystem}); grantErr != nil {
					return grantErr
				}
			}
			if err := service.store.MarkPaydayClaimed(txCtx, payday.ID); err != nil {
				return err
			}
			claimed = append(claimed, payday)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if service.events != nil {
		for _, payday := range claimed {
			_ = service.events.Publish(ctx, bus.Event{Name: paydayevent.Name, Payload: paydayevent.Payload{PlayerID: playerID, Amount: payday.TotalAwarded, CurrencyType: payday.CurrencyType, Streak: payday.StreakDays}})
		}
	}

	return nil
}
