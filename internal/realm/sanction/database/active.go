package database

import (
	"context"
	"time"

	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
)

// Active returns the current hot-path projection for one player.
func (repository *Repository) Active(ctx context.Context, playerID int64, now time.Time) (sanctionrecord.ActiveState, error) {
	rows, err := repository.executor(ctx).Query(ctx, punishmentSelect+` where receiver_player_id=$1 and revoked_at is null and (expires_at is null or expires_at>$2) and kind in ('ban','mute','trade_lock') order by issued_at desc`, playerID, now)
	if err != nil {
		return sanctionrecord.ActiveState{}, err
	}
	defer rows.Close()
	var state sanctionrecord.ActiveState
	for rows.Next() {
		value, scanErr := scanPunishment(rows)
		if scanErr != nil {
			return state, scanErr
		}
		switch value.Kind {
		case sanctionrecord.KindBan:
			if state.Ban == nil {
				copy := value
				state.Ban = &copy
			}
		case sanctionrecord.KindMute:
			state.MutedPermanently, state.MuteUntil = mergeExpiry(state.MutedPermanently, state.MuteUntil, value.ExpiresAt)
		case sanctionrecord.KindTradeLock:
			state.TradeLockedPermanently, state.TradeLockUntil = mergeExpiry(state.TradeLockedPermanently, state.TradeLockUntil, value.ExpiresAt)
		}
	}
	return state, rows.Err()
}

// mergeExpiry merges overlapping active expiry windows.
func mergeExpiry(permanent bool, current *time.Time, candidate *time.Time) (bool, *time.Time) {
	if permanent || candidate == nil {
		return true, nil
	}
	if current == nil || candidate.After(*current) {
		value := *candidate
		return false, &value
	}
	return false, current
}

// Revoke atomically marks one active punishment revoked.
func (repository *Repository) Revoke(ctx context.Context, id int64, actorID *int64, now time.Time) (sanctionrecord.Punishment, bool, error) {
	rows, err := repository.executor(ctx).Query(ctx, `update punishments set revoked_at=$2,revoked_by_player_id=$3 where id=$1 and revoked_at is null returning id,receiver_player_id,issuer_player_id,issuer_kind,kind,reason,cfh_topic_id,issue_id,source,issued_at,expires_at,revoked_at,revoked_by_player_id`, id, now, actorID)
	if err != nil {
		return sanctionrecord.Punishment{}, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return sanctionrecord.Punishment{}, false, rows.Err()
	}
	value, err := scanPunishment(rows)
	return value, err == nil, err
}
