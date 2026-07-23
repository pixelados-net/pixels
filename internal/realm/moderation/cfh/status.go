package cfh

import (
	"context"
	"time"

	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/networking/codec"
	outsanction "github.com/niflaot/pixels/networking/outbound/moderation/cfh/sanctionstatus"
)

// sanctionStatus derives Nitro's panel fields from common punishment truth.
func (handler Handler) SanctionStatus(ctx context.Context, playerID int64) (codec.Packet, error) {
	now := time.Now()
	state, err := handler.Sanctions.Active(ctx, playerID)
	if err != nil {
		return codec.Packet{}, err
	}
	ladder, err := handler.Sanctions.Store().Ladder(ctx)
	if err != nil {
		return codec.Packet{}, err
	}
	params := outsanction.Params{Name: "ALERT", Reason: "cfh.reason.EMPTY", NextName: "ALERT"}
	current, level, found, err := handler.Sanctions.Store().LastEscalation(ctx, playerID)
	if err != nil {
		return codec.Packet{}, err
	}
	nextLevel := int32(1)
	if found {
		params.Active, params.Reason, params.CreatedAt = current.ActiveAt(now), current.Reason, current.IssuedAt.Format(time.RFC3339)
		params.Name = sanctionWireName(current.Kind, current.ExpiresAt)
		if current.ExpiresAt != nil {
			params.LengthHours = int32(current.ExpiresAt.Sub(current.IssuedAt).Hours())
		}
		if entry, ok := sanctionLadderEntry(ladder, level); ok {
			probationEnd := current.IssuedAt.Add(time.Duration(entry.ProbationDays) * 24 * time.Hour)
			if probationEnd.After(now) {
				params.ProbationHoursLeft = int32(probationEnd.Sub(now).Hours())
				nextLevel = level + 1
			}
		}
	}
	if next, ok := sanctionLadderEntry(ladder, nextLevel); ok {
		params.NextName = sanctionWireName(next.Kind, ladderExpiry(next))
		params.NextLengthHours = next.DurationHours
	}
	params.HasCustomMute = state.MutedPermanently || state.MuteUntil != nil
	if state.TradeLockUntil != nil {
		params.TradeLockExpiry = state.TradeLockUntil.Format(time.RFC3339)
	} else if state.TradeLockedPermanently {
		params.TradeLockExpiry = "PERMANENT"
	}
	return outsanction.Encode(params)
}

// sanctionLadderEntry resolves a requested level or the final configured level.
func sanctionLadderEntry(entries []sanctionrecord.LadderEntry, level int32) (sanctionrecord.LadderEntry, bool) {
	if len(entries) == 0 {
		return sanctionrecord.LadderEntry{}, false
	}
	for _, entry := range entries {
		if entry.Level >= level {
			return entry, true
		}
	}
	return entries[len(entries)-1], true
}

// ladderExpiry creates a non-nil marker for finite Nitro ban naming.
func ladderExpiry(entry sanctionrecord.LadderEntry) *time.Time {
	if entry.DurationHours <= 0 {
		return nil
	}
	value := time.Now().Add(time.Duration(entry.DurationHours) * time.Hour)
	return &value
}

// sanctionWireName maps internal kinds to Nitro translation switches.
func sanctionWireName(kind sanctionrecord.Kind, expiresAt *time.Time) string {
	switch kind {
	case sanctionrecord.KindMute:
		return "MUTE"
	case sanctionrecord.KindBan:
		if expiresAt == nil {
			return "BAN_PERMANENT"
		}
		return "BAN"
	default:
		return "ALERT"
	}
}
