// Package effect implements PostgreSQL player effect persistence.
package effect

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository persists player effects in PostgreSQL.
type Repository struct {
	// pool owns transaction scopes.
	pool *postgres.Pool
}

// New creates an effect repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{pool: pool}
}

// WithinTransaction runs work in a shared PostgreSQL transaction.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, active := postgres.ScopedExecutor(ctx); active {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

// List returns one player's effects.
func (repository *Repository) List(ctx context.Context, playerID int64) ([]playereffect.Effect, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.pool).Query(ctx, `select player_id,effect_id,duration_seconds,activated_at,remaining_charges from player_effects where player_id=$1 order by effect_id`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player effects: %w", err)
	}
	defer rows.Close()
	effects := make([]playereffect.Effect, 0)
	for rows.Next() {
		effect, scanErr := scanEffect(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		effects = append(effects, effect)
	}
	return effects, rows.Err()
}

// Grant creates or increments one effect stack.
func (repository *Repository) Grant(ctx context.Context, playerID int64, effectID int32, durationSeconds int32) (playereffect.Effect, error) {
	row := postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, `insert into player_effects(player_id,effect_id,duration_seconds) values($1,$2,$3) on conflict(player_id,effect_id) do update set duration_seconds=excluded.duration_seconds,remaining_charges=least(99,player_effects.remaining_charges+1),updated_at=now() returning player_id,effect_id,duration_seconds,activated_at,remaining_charges`, playerID, effectID, durationSeconds)
	effect, err := scanEffect(row)
	if err != nil {
		return playereffect.Effect{}, fmt.Errorf("grant player effect: %w", err)
	}
	return effect, nil
}

// Activate starts one available effect charge.
func (repository *Repository) Activate(ctx context.Context, playerID int64, effectID int32, now time.Time) (playereffect.Effect, bool, error) {
	row := postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, `update player_effects set activated_at=coalesce(activated_at,$3),updated_at=now() where player_id=$1 and effect_id=$2 and remaining_charges>0 returning player_id,effect_id,duration_seconds,activated_at,remaining_charges`, playerID, effectID, now)
	effect, err := scanEffect(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return playereffect.Effect{}, false, nil
	}
	if err != nil {
		return playereffect.Effect{}, false, fmt.Errorf("activate player effect: %w", err)
	}
	return effect, true, nil
}

// SetActive replaces the selected effect id.
func (repository *Repository) SetActive(ctx context.Context, playerID int64, effectID *int32) error {
	_, err := postgres.ExecutorFor(ctx, repository.pool).Exec(ctx, `update players set active_effect_id=$2,updated_at=now(),version=version+1 where id=$1 and deleted_at is null`, playerID, effectID)
	return err
}

// Active returns one player's selected effect id.
func (repository *Repository) Active(ctx context.Context, playerID int64) (*int32, error) {
	var effectID *int32
	err := postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, `select active_effect_id from players where id=$1 and deleted_at is null`, playerID).Scan(&effectID)
	return effectID, err
}

// Revoke deletes one effect stack.
func (repository *Repository) Revoke(ctx context.Context, playerID int64, effectID int32) (bool, error) {
	result, err := postgres.ExecutorFor(ctx, repository.pool).Exec(ctx, `delete from player_effects where player_id=$1 and effect_id=$2`, playerID, effectID)
	return result.RowsAffected() > 0, err
}

// Expire consumes one charge from each expired stack in a bounded locked batch.
func (repository *Repository) Expire(ctx context.Context, now time.Time, limit int32) ([]playereffect.Expiration, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.pool).Query(ctx, `with selected as materialized (select player_id,effect_id,remaining_charges from player_effects where activated_at is not null and duration_seconds>0 and activated_at+make_interval(secs=>duration_seconds)<=$1 order by activated_at for update skip locked limit $2), cleared as (update players p set active_effect_id=null,updated_at=now(),version=version+1 from selected s where p.id=s.player_id and p.active_effect_id=s.effect_id returning p.id as player_id,s.effect_id), deleted as (delete from player_effects e using selected s where e.player_id=s.player_id and e.effect_id=s.effect_id and s.remaining_charges=1 returning e.player_id,e.effect_id,0::integer as remaining_charges), updated as (update player_effects e set remaining_charges=e.remaining_charges-1,activated_at=null,updated_at=now() from selected s where e.player_id=s.player_id and e.effect_id=s.effect_id and s.remaining_charges>1 returning e.player_id,e.effect_id,e.remaining_charges), changed as (select * from deleted union all select * from updated) select c.player_id,c.effect_id,c.remaining_charges,exists(select 1 from cleared x where x.player_id=c.player_id and x.effect_id=c.effect_id) from changed c`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	expired := make([]playereffect.Expiration, 0)
	for rows.Next() {
		var item playereffect.Expiration
		if scanErr := rows.Scan(&item.PlayerID, &item.EffectID, &item.RemainingCharges, &item.Selected); scanErr != nil {
			return nil, scanErr
		}
		expired = append(expired, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return expired, nil
}

// scanEffect scans one effect row.
func scanEffect(row interface{ Scan(...any) error }) (playereffect.Effect, error) {
	var effect playereffect.Effect
	err := row.Scan(&effect.PlayerID, &effect.ID, &effect.DurationSeconds, &effect.ActivatedAt, &effect.RemainingCharges)
	return effect, err
}

// storeAssertion verifies Repository implements the effect store.
var storeAssertion playereffect.Store = (*Repository)(nil)
