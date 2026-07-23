// Package achievement implements player achievement persistence in PostgreSQL.
package achievement

import (
	"context"

	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository persists player achievements.
type Repository struct {
	// pool owns transaction scopes.
	pool *postgres.Pool
}

// New creates an achievement repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// Badges lists a player's durable badge snapshot.
func (repository *Repository) Badges(ctx context.Context, playerID int64) ([]playerachievement.Badge, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.pool).Query(ctx, `select id,code,equipped,coalesce(slot,0) from player_badges where player_id=$1 order by slot nulls last,code`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	badges := make([]playerachievement.Badge, 0)
	for rows.Next() {
		var badge playerachievement.Badge
		if err = rows.Scan(&badge.ID, &badge.Code, &badge.Equipped, &badge.Slot); err != nil {
			return nil, err
		}
		badges = append(badges, badge)
	}
	return badges, rows.Err()
}

// SetEquipped atomically replaces one player's active badge slots.
func (repository *Repository) SetEquipped(ctx context.Context, playerID int64, codes []string) error {
	_, err := postgres.ExecutorFor(ctx, repository.pool).Exec(ctx, `update player_badges set equipped=array_position($2::text[],upper(code)) is not null,slot=array_position($2::text[],upper(code))::smallint where player_id=$1`, playerID, codes)
	return err
}

// GrantBadge grants one badge idempotently.
func (repository *Repository) GrantBadge(ctx context.Context, playerID int64, code string, source string) (bool, error) {
	result, err := postgres.ExecutorFor(ctx, repository.pool).Exec(ctx, `insert into player_badges(player_id,code,source) values($1,$2,$3) on conflict do nothing`, playerID, code, source)
	return result.RowsAffected() > 0, err
}

// ReplaceBadge replaces one badge code while preserving its equipped slot.
func (repository *Repository) ReplaceBadge(ctx context.Context, playerID int64, oldCode string, newCode string, source string) (bool, error) {
	result, err := postgres.ExecutorFor(ctx, repository.pool).Exec(ctx, `update player_badges set code=$3,source=$4 where player_id=$1 and upper(code)=upper($2) and not exists(select 1 from player_badges where player_id=$1 and upper(code)=upper($3))`, playerID, oldCode, newCode, source)
	return result.RowsAffected() > 0, err
}

// RemoveBadge removes one badge regardless of equipped state.
func (repository *Repository) RemoveBadge(ctx context.Context, playerID int64, code string) (bool, error) {
	result, err := postgres.ExecutorFor(ctx, repository.pool).Exec(ctx, `delete from player_badges where player_id=$1 and upper(code)=upper($2)`, playerID, code)
	return result.RowsAffected() > 0, err
}

// GrantRespect applies one idempotent positive respect grant.
func (repository *Repository) GrantRespect(ctx context.Context, playerID int64, amount int32, sourceKey string, source string) (bool, error) {
	granted := false
	work := func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		result, err := executor.Exec(txCtx, `insert into player_respect_ledger(source_key,player_id,amount,source) values($1,$2,$3,$4) on conflict do nothing`, sourceKey, playerID, amount, source)
		if err != nil || result.RowsAffected() == 0 {
			return err
		}
		granted = true
		_, err = executor.Exec(txCtx, `insert into player_respect_totals(player_id,received) values($1,$2) on conflict(player_id) do update set received=player_respect_totals.received+excluded.received,updated_at=now()`, playerID, amount)
		return err
	}
	if _, active := postgres.ScopedExecutor(ctx); active {
		return granted, work(ctx)
	}
	err := postgres.WithinScope(ctx, repository.pool, work)
	return granted, err
}
