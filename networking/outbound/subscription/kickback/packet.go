// Package kickback contains the SCR_SEND_KICKBACK_INFO outbound packet.
package kickback

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies SCR_SEND_KICKBACK_INFO.
	Header uint16 = 3277
)

// Info contains kickback summary data.
type Info struct {
	// Streak stores consecutive HC days.
	Streak int32
	// FirstSubscriptionDate stores the first activation date.
	FirstSubscriptionDate string
	// Percentage stores the configured kickback fraction.
	Percentage float64
	// CreditsMissed stores historical missed credits.
	CreditsMissed int32
	// CreditsRewarded stores historical rewarded credits.
	CreditsRewarded int32
	// CreditsSpent stores eligible period spending.
	CreditsSpent int32
	// StreakBonus stores the current streak bonus.
	StreakBonus int32
	// MonthlyBonus stores the current spending bonus.
	MonthlyBonus int32
	// MinutesUntilPayday stores Nitro's next-cycle countdown unit.
	MinutesUntilPayday int32
}

// Encode creates a SCR_SEND_KICKBACK_INFO packet.
func Encode(info Info) (codec.Packet, error) {
	definition := codec.Definition{codec.Int32Field, codec.StringField, codec.DoubleField, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	return codec.NewPacket(Header, definition, codec.Int32(info.Streak), codec.String(info.FirstSubscriptionDate),
		codec.Float64(info.Percentage), codec.Int32(info.CreditsMissed), codec.Int32(info.CreditsRewarded),
		codec.Int32(info.CreditsSpent), codec.Int32(info.StreakBonus), codec.Int32(info.MonthlyBonus), codec.Int32(info.MinutesUntilPayday))
}
