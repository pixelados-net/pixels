package club

import (
	"context"
	"math"
	"time"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstatus "github.com/niflaot/pixels/networking/outbound/subscription/status"
)

// statusProjection maps durable membership state to Nitro fields.
func statusProjection(membership record.Membership, responseType int32, now time.Time, requestedProduct string) outstatus.State {
	remaining := time.Duration(0)
	if membership.ExpiresAt != nil && membership.ExpiresAt.After(now) {
		remaining = membership.ExpiresAt.Sub(now)
	}
	days := int32(math.Ceil(remaining.Hours() / 24))
	minutes := int32(remaining / time.Minute)
	product := "habbo_club"
	if requestedProduct != "" {
		product = requestedProduct
	}
	return outstatus.State{ProductName: product, DaysToPeriodEnd: days,
		MemberPeriods: int32(membership.LifetimeActiveSeconds / corePeriodSeconds), ResponseType: responseType,
		EverMember: membership.StartedAt != nil, VIP: membership.Level == record.LevelVIP,
		PastClubDays: int32(membership.LifetimeActiveSeconds / 86_400),
		PastVIPDays:  int32(membership.LifetimeVIPSeconds / 86_400), MinutesUntilExpiration: minutes}
}

const (
	// corePeriodSeconds stores one 31-day club period.
	corePeriodSeconds int64 = 2_678_400
)

// clubProjection maps membership to the player entitlement cache.
func clubProjection(membership record.Membership) playermodel.Club {
	return playermodel.Club{Level: playermodel.ClubLevel(membership.Level), ExpiresAt: membership.ExpiresAt}
}

// clampInt32 clamps database amounts to protocol range.
func clampInt32(value int64) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}
	if value < math.MinInt32 {
		return math.MinInt32
	}

	return int32(value)
}

// send sends one encoded club packet.
func send(ctx context.Context, connection netconn.Context, packet codec.Packet, err error) error {
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
