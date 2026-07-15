package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// LastEscalation returns the most recent escalated punishment and its ladder level.
func (repository *Repository) LastEscalation(ctx context.Context, playerID int64) (sanctionrecord.Punishment, int32, bool, error) {
	rows, err := repository.executor(ctx).Query(ctx, punishmentSelect+` where receiver_player_id=$1 and source in ('escalation','cfh_auto') order by issued_at desc limit 1`, playerID)
	if err != nil {
		return sanctionrecord.Punishment{}, 0, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return sanctionrecord.Punishment{}, 0, false, rows.Err()
	}
	value, err := scanPunishment(rows)
	if err != nil {
		return sanctionrecord.Punishment{}, 0, false, err
	}
	var level int32
	err = repository.executor(ctx).QueryRow(ctx, `select level from sanction_ladder where kind=$1 and duration_hours=case when $2::timestamptz is null then 0 else greatest(1,ceil(extract(epoch from ($2-$3))/3600)::integer) end order by level desc limit 1`, value.Kind, value.ExpiresAt, value.IssuedAt).Scan(&level)
	if errors.Is(err, pgx.ErrNoRows) {
		level = 1
	} else if err != nil {
		return sanctionrecord.Punishment{}, 0, false, err
	}
	return value, level, true, nil
}

// Ladder returns escalation policy ordered by level.
func (repository *Repository) Ladder(ctx context.Context) ([]sanctionrecord.LadderEntry, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select level,kind,duration_hours,probation_days from sanction_ladder order by level`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]sanctionrecord.LadderEntry, 0)
	for rows.Next() {
		var value sanctionrecord.LadderEntry
		if err = rows.Scan(&value.Level, &value.Kind, &value.DurationHours, &value.ProbationDays); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// ReplaceLadder atomically replaces escalation policy.
func (repository *Repository) ReplaceLadder(ctx context.Context, entries []sanctionrecord.LadderEntry) error {
	return postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		if _, err := repository.executor(txCtx).Exec(txCtx, `delete from sanction_ladder`); err != nil {
			return err
		}
		for _, entry := range entries {
			if _, err := repository.executor(txCtx).Exec(txCtx, `insert into sanction_ladder(level,kind,duration_hours,probation_days) values($1,$2,$3,$4)`, entry.Level, entry.Kind, entry.DurationHours, entry.ProbationDays); err != nil {
				return err
			}
		}
		return nil
	})
}
