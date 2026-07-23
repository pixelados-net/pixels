// Package record contains durable subscription data.
package record

import "time"

// Level identifies one club entitlement tier.
type Level int16

const (
	// LevelNone identifies no active entitlement.
	LevelNone Level = iota
	// LevelHC identifies Habbo Club.
	LevelHC
	// LevelVIP identifies the highest club tier.
	LevelVIP
)

// Membership contains one player's subscription lifecycle.
type Membership struct {
	// PlayerID identifies the player.
	PlayerID int64
	// Level stores the current tier.
	Level Level
	// StartedAt stores the first activation instant.
	StartedAt *time.Time
	// StreakStartedAt stores the current uninterrupted membership start.
	StreakStartedAt *time.Time
	// ExpiresAt stores the exclusive entitlement expiration.
	ExpiresAt *time.Time
	// LastPaydayAt stores the last evaluated kickback instant.
	LastPaydayAt *time.Time
	// LastAccruedAt stores the last durable active-time boundary.
	LastAccruedAt *time.Time
	// LifetimeActiveSeconds stores accumulated active club time.
	LifetimeActiveSeconds int64
	// LifetimeVIPSeconds stores accumulated time at the highest tier.
	LifetimeVIPSeconds int64
	// GiftsEarned stores materialized monthly gift periods.
	GiftsEarned int32
	// GiftsClaimed stores claimed monthly gifts.
	GiftsClaimed int32
	// Version stores the optimistic version.
	Version int64
}

// Offer contains one purchasable club duration.
type Offer struct {
	// ID identifies the offer.
	ID int64
	// Name stores the stable product code.
	Name string
	// DayCount stores granted subscription days.
	DayCount int32
	// PriceCredits stores the credits price.
	PriceCredits int64
	// PricePoints stores the activity-points price.
	PricePoints int64
	// PointsType identifies the points currency.
	PointsType int32
	// VIP reports whether the offer grants VIP.
	VIP bool
	// Deal reports whether the offer is extension-only.
	Deal bool
	// Enabled reports whether the offer can be purchased.
	Enabled bool
	// OrderNum stores stable display order.
	OrderNum int32
}

// Payday contains one durable kickback reward.
type Payday struct {
	// ID identifies the reward.
	ID int64
	// PlayerID identifies the beneficiary.
	PlayerID int64
	// OccurredAt stores the evaluation instant.
	OccurredAt time.Time
	// StreakDays stores active subscription days.
	StreakDays int32
	// CreditsSpent stores eligible catalog spend.
	CreditsSpent int64
	// StreakBonus stores the streak reward.
	StreakBonus int64
	// MonthlyBonus stores the spending kickback.
	MonthlyBonus int64
	// TotalAwarded stores the total reward.
	TotalAwarded int64
	// CurrencyType identifies the reward currency.
	CurrencyType int32
	// Claimed reports whether currency was delivered.
	Claimed bool
}
