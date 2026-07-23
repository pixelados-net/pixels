// Package profile implements PostgreSQL public-profile persistence.
package profile

import (
	"context"
	"fmt"
	"time"

	playerprofile "github.com/niflaot/pixels/internal/realm/player/profile"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository persists public tags and daily respect grants.
type Repository struct {
	// pool owns transaction scopes.
	pool *postgres.Pool
}

// New creates a public-profile repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// Tags returns ordered public tags.
func (repository *Repository) Tags(ctx context.Context, playerID int64) ([]string, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.pool).Query(ctx, `select tag from player_profile_tags where player_id=$1 order by position`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player profile tags: %w", err)
	}
	defer rows.Close()
	tags := make([]string, 0, playerprofile.MaxTags)
	for rows.Next() {
		var tag string
		if err = rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("scan player profile tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// ReplaceTags atomically replaces ordered public tags.
func (repository *Repository) ReplaceTags(ctx context.Context, playerID int64, tags []string) error {
	return postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		if _, err := executor.Exec(txCtx, `delete from player_profile_tags where player_id=$1`, playerID); err != nil {
			return err
		}
		for index, tag := range tags {
			if _, err := executor.Exec(txCtx, `insert into player_profile_tags(player_id,position,tag) values($1,$2,$3)`, playerID, index+1, tag); err != nil {
				return err
			}
		}
		return nil
	})
}

// RespectState returns total and remaining user and pet respect allowances.
func (repository *Repository) RespectState(ctx context.Context, playerID int64, date time.Time, userLimit int, petLimit int) (playerprofile.RespectState, error) {
	var state playerprofile.RespectState
	var userUsed int32
	var petUsed int32
	err := postgres.ExecutorFor(ctx, repository.pool).QueryRow(ctx, `select coalesce((select received from player_respect_totals where player_id=$1),0),coalesce((select count(*) from player_respect_grants where actor_player_id=$1 and grant_date=$2),0),coalesce((select count(*) from pet_respects where actor_player_id=$1 and respected_on=$2),0)`, playerID, date.Format("2006-01-02")).Scan(&state.Received, &userUsed, &petUsed)
	if err != nil {
		return playerprofile.RespectState{}, fmt.Errorf("read player respect state: %w", err)
	}
	state.UserRemaining = remaining(userLimit, userUsed)
	state.PetRemaining = remaining(petLimit, petUsed)
	return state, nil
}

// GrantRespect serializes and applies one daily user respect.
func (repository *Repository) GrantRespect(ctx context.Context, actorID int64, targetID int64, date time.Time, limit int, unlimited bool) (result playerprofile.RespectResult, err error) {
	err = postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		dateValue := date.Format("2006-01-02")
		if _, lockErr := executor.Exec(txCtx, `select pg_advisory_xact_lock(hashtextextended('user-respect:'||$1::text||':'||$2::text,0))`, actorID, dateValue); lockErr != nil {
			return lockErr
		}
		var used int32
		if countErr := executor.QueryRow(txCtx, `select count(*) from player_respect_grants where actor_player_id=$1 and grant_date=$2`, actorID, dateValue).Scan(&used); countErr != nil {
			return countErr
		}
		result.Remaining = remaining(limit, used)
		if !unlimited && result.Remaining == 0 {
			return nil
		}
		command, insertErr := executor.Exec(txCtx, `insert into player_respect_grants(actor_player_id,target_player_id,grant_date,source) values($1,$2,$3,'user') on conflict do nothing`, actorID, targetID, dateValue)
		if insertErr != nil {
			return insertErr
		}
		if command.RowsAffected() == 0 {
			result.Duplicate = true
			return nil
		}
		sourceKey := fmt.Sprintf("user:%d:%d:%s", actorID, targetID, dateValue)
		if _, insertErr = executor.Exec(txCtx, `insert into player_respect_ledger(source_key,player_id,amount,source) values($1,$2,1,'user') on conflict do nothing`, sourceKey, targetID); insertErr != nil {
			return insertErr
		}
		if scanErr := executor.QueryRow(txCtx, `insert into player_respect_totals(player_id,received) values($1,1) on conflict(player_id) do update set received=player_respect_totals.received+1,updated_at=now() returning received`, targetID).Scan(&result.TotalReceived); scanErr != nil {
			return scanErr
		}
		result.Applied = true
		if unlimited {
			result.Remaining = int32(limit)
		} else {
			result.Remaining--
		}
		return nil
	})
	return result, err
}

// remaining computes a bounded daily allowance.
func remaining(limit int, used int32) int32 {
	value := int32(limit) - used
	if value < 0 {
		return 0
	}
	return value
}
