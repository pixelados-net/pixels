package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// FindMembership finds one membership and optionally locks it.
func (repository *Repository) FindMembership(ctx context.Context, playerID int64, lock bool) (record.Membership, bool, error) {
	query := `select player_id,level,started_at,streak_started_at,expires_at,last_payday_at,last_accrued_at,lifetime_active_seconds,lifetime_vip_seconds,gifts_earned,gifts_claimed,version from subscription_memberships where player_id=$1`
	if lock {
		query += ` for update`
	}
	membership, err := scanMembership(repository.executorFor(ctx).QueryRow(ctx, query, playerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return record.Membership{}, false, nil
	}
	return membership, err == nil, err
}

// UpsertMembership writes one membership.
func (repository *Repository) UpsertMembership(ctx context.Context, membership record.Membership) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into subscription_memberships (player_id,level,started_at,streak_started_at,expires_at,last_payday_at,last_accrued_at,lifetime_active_seconds,lifetime_vip_seconds,gifts_earned,gifts_claimed,version) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,1) on conflict (player_id) do update set level=excluded.level,started_at=coalesce(subscription_memberships.started_at,excluded.started_at),streak_started_at=excluded.streak_started_at,expires_at=excluded.expires_at,last_payday_at=excluded.last_payday_at,last_accrued_at=excluded.last_accrued_at,lifetime_active_seconds=excluded.lifetime_active_seconds,lifetime_vip_seconds=excluded.lifetime_vip_seconds,gifts_earned=excluded.gifts_earned,gifts_claimed=excluded.gifts_claimed,version=subscription_memberships.version+1`, membership.PlayerID, membership.Level, membership.StartedAt, membership.StreakStartedAt, membership.ExpiresAt, membership.LastPaydayAt, membership.LastAccruedAt, membership.LifetimeActiveSeconds, membership.LifetimeVIPSeconds, membership.GiftsEarned, membership.GiftsClaimed)
	return err
}

// ListDueMemberships lists memberships crossing one durable lifecycle boundary.
func (repository *Repository) ListDueMemberships(ctx context.Context, now time.Time, paydayInterval time.Duration, giftPeriodSeconds int64) ([]record.Membership, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select player_id,level,started_at,streak_started_at,expires_at,last_payday_at,last_accrued_at,lifetime_active_seconds,lifetime_vip_seconds,gifts_earned,gifts_claimed,version from subscription_memberships where level>0 and (expires_at is null or expires_at<=$1 or last_payday_at is null or last_payday_at+($2*interval '1 second')<=least($1,expires_at) or last_accrued_at is null or lifetime_active_seconds+greatest(0,extract(epoch from least($1,expires_at)-last_accrued_at)::bigint)>=(gifts_earned::bigint+1)*$3) order by player_id`, now, int64(paydayInterval/time.Second), giftPeriodSeconds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]record.Membership, 0)
	for rows.Next() {
		membership, scanErr := scanMembership(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, membership)
	}
	return result, rows.Err()
}

// scanMembership scans one membership.
func scanMembership(row pgx.Row) (record.Membership, error) {
	var membership record.Membership
	var started, streak, expires, payday, accrued pgtype.Timestamptz
	err := row.Scan(&membership.PlayerID, &membership.Level, &started, &streak, &expires, &payday, &accrued, &membership.LifetimeActiveSeconds, &membership.LifetimeVIPSeconds, &membership.GiftsEarned, &membership.GiftsClaimed, &membership.Version)
	if started.Valid {
		membership.StartedAt = &started.Time
	}
	if streak.Valid {
		membership.StreakStartedAt = &streak.Time
	}
	if expires.Valid {
		membership.ExpiresAt = &expires.Time
	}
	if payday.Valid {
		membership.LastPaydayAt = &payday.Time
	}
	if accrued.Valid {
		membership.LastAccruedAt = &accrued.Time
	}
	return membership, err
}

// InsertPayday writes one kickback reward.
func (repository *Repository) InsertPayday(ctx context.Context, payday record.Payday) (record.Payday, error) {
	err := repository.executorFor(ctx).QueryRow(ctx, `insert into subscription_payday_log (player_id,occurred_at,streak_days,credits_spent,streak_bonus,monthly_bonus,total_awarded,currency_type,claimed) values ($1,$2,$3,$4,$5,$6,$7,$8,$9) returning id,occurred_at`, payday.PlayerID, payday.OccurredAt, payday.StreakDays, payday.CreditsSpent, payday.StreakBonus, payday.MonthlyBonus, payday.TotalAwarded, payday.CurrencyType, payday.Claimed).Scan(&payday.ID, &payday.OccurredAt)
	return payday, err
}

// ListUnclaimedPaydays lists pending rewards.
func (repository *Repository) ListUnclaimedPaydays(ctx context.Context, playerID int64) ([]record.Payday, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id,player_id,occurred_at,streak_days,credits_spent,streak_bonus,monthly_bonus,total_awarded,currency_type,claimed from subscription_payday_log where player_id=$1 and not claimed order by id for update`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]record.Payday, 0)
	for rows.Next() {
		var payday record.Payday
		if err := rows.Scan(&payday.ID, &payday.PlayerID, &payday.OccurredAt, &payday.StreakDays, &payday.CreditsSpent, &payday.StreakBonus, &payday.MonthlyBonus, &payday.TotalAwarded, &payday.CurrencyType, &payday.Claimed); err != nil {
			return nil, fmt.Errorf("scan subscription payday: %w", err)
		}
		result = append(result, payday)
	}
	return result, rows.Err()
}

// MarkPaydayClaimed marks one reward delivered.
func (repository *Repository) MarkPaydayClaimed(ctx context.Context, id int64) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `update subscription_payday_log set claimed=true where id=$1 and not claimed`, id)
	return err
}

// InsertGiftClaim records one monthly gift claim.
func (repository *Repository) InsertGiftClaim(ctx context.Context, playerID int64, period time.Time, itemID int64) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into subscription_club_gift_claims (player_id,period_start,claimed_item_id) values ($1,$2,$3)`, playerID, period, itemID)
	return err
}
