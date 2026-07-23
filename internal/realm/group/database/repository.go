// Package database implements social group reads in PostgreSQL.
package database

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	forumdb "github.com/niflaot/pixels/internal/realm/group/database/forum"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository reads social group membership.
type Repository struct {
	// pool executes PostgreSQL operations.
	pool *postgres.Pool
	// Repository promotes forum-specific PostgreSQL behavior.
	*forumdb.Repository
}

// New creates a social group repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{pool: pool, Repository: forumdb.New(pool)}
}

// executor returns the active transaction or repository pool.
func (repository *Repository) executor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.pool)
}

// WithinTransaction runs work atomically in a shared transaction scope.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, scoped := postgres.ScopedExecutor(ctx); scoped {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

// RoomMembership returns the linked group and current member IDs.
func (repository *Repository) RoomMembership(ctx context.Context, roomID int64) (int64, []int64, bool, error) {
	var groupID int64
	err := repository.pool.QueryRow(ctx, `select group_id from room_social_groups where room_id=$1`, roomID).Scan(&groupID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil, false, nil
	}
	if err != nil {
		return 0, nil, false, err
	}
	rows, err := repository.pool.Query(ctx, `select player_id from social_group_members where group_id=$1 order by player_id`, groupID)
	if err != nil {
		return 0, nil, false, err
	}
	defer rows.Close()
	players := make([]int64, 0)
	for rows.Next() {
		var playerID int64
		if err = rows.Scan(&playerID); err != nil {
			return 0, nil, false, err
		}
		players = append(players, playerID)
	}
	return groupID, players, true, rows.Err()
}

// audit writes one attributed mutation inside the current transaction when requested.
func (repository *Repository) audit(ctx context.Context, groupID int64, action string, targetPlayerID int64, version int64) error {
	attribution, found := grouprecord.AuditFromContext(ctx)
	if !found {
		return nil
	}
	var target any
	if targetPlayerID > 0 {
		target = targetPlayerID
	}
	var storedVersion any
	if version > 0 {
		storedVersion = version
	}
	_, err := repository.executor(ctx).Exec(ctx, `insert into social_group_audit(group_id,actor_player_id,action,target_player_id,reason,version) values($1,$2,$3,$4,$5,$6)`, groupID, attribution.ActorPlayerID, action, target, strings.TrimSpace(attribution.Reason), storedVersion)
	return err
}

// ClaimCreateOperation claims or replays one administrative creation key.
func (repository *Repository) ClaimCreateOperation(ctx context.Context, key string, requestHash string) (int64, bool, error) {
	attribution, found := grouprecord.AuditFromContext(ctx)
	if !found {
		return 0, false, grouprecord.ErrForbidden
	}
	command, err := repository.executor(ctx).Exec(ctx, `insert into social_group_create_operations(operation_key,actor_player_id,request_hash) values($1,$2,$3) on conflict do nothing`, key, attribution.ActorPlayerID, requestHash)
	if err != nil {
		return 0, false, mapConflict(err)
	}
	if command.RowsAffected() == 1 {
		return 0, false, nil
	}
	var actorID int64
	var storedHash string
	var groupID *int64
	if err = repository.executor(ctx).QueryRow(ctx, `select actor_player_id,request_hash,group_id from social_group_create_operations where operation_key=$1 for update`, key).Scan(&actorID, &storedHash, &groupID); err != nil {
		return 0, false, err
	}
	if actorID != attribution.ActorPlayerID || storedHash != requestHash || groupID == nil {
		return 0, false, grouprecord.ErrConflict
	}
	return *groupID, true, nil
}

// CompleteCreateOperation binds a claimed creation key to its committed group.
func (repository *Repository) CompleteCreateOperation(ctx context.Context, key string, groupID int64) error {
	command, err := repository.executor(ctx).Exec(ctx, `update social_group_create_operations set group_id=$2,completed_at=now() where operation_key=$1 and group_id is null`, key, groupID)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return grouprecord.ErrConflict
	}
	return nil
}
