// Package votes contains PostgreSQL room vote persistence.
package votes

import (
	"context"
	"fmt"

	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// insertVoteSQL inserts one unique room vote.
	insertVoteSQL = `INSERT INTO room_votes (room_id, player_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	// incrementScoreSQL increments and returns one active room score.
	incrementScoreSQL = `UPDATE rooms SET score = score + 1, updated_at = NOW(), version = version + 1 WHERE id = $1 AND deleted_at IS NULL RETURNING score`
	// selectScoreSQL reads one active room score.
	selectScoreSQL = `SELECT score FROM rooms WHERE id = $1 AND deleted_at IS NULL`
	// hasVoteSQL tests one unique room vote.
	hasVoteSQL = `SELECT EXISTS (SELECT 1 FROM room_votes WHERE room_id = $1 AND player_id = $2)`
	// existingVotesSQL selects voters from one bounded active set.
	existingVotesSQL = `SELECT player_id FROM room_votes WHERE room_id = $1 AND player_id = ANY($2::bigint[])`
	// listVotesSQL lists votes with optional player and time filters.
	listVotesSQL = `SELECT room_id, player_id, created_at FROM room_votes WHERE room_id = $1 AND ($2::bigint IS NULL OR player_id = $2) AND ($3::timestamptz IS NULL OR created_at < $3) ORDER BY created_at DESC, player_id DESC LIMIT $4`
)

// transactionRunner runs repository work atomically.
type transactionRunner func(context.Context, func(context.Context, postgres.Executor) error) error

// Repository persists room votes.
type Repository struct {
	// executor runs PostgreSQL statements.
	executor postgres.Executor
	// withinTx runs atomic vote mutations.
	withinTx transactionRunner
}

// New creates a room vote repository.
func New(executor postgres.Executor) *Repository {
	repository := &Repository{executor: executor}
	pool, ok := executor.(*postgres.Pool)
	if ok {
		repository.withinTx = func(ctx context.Context, work func(context.Context, postgres.Executor) error) error {
			return postgres.WithinScope(ctx, pool, func(txCtx context.Context) error {
				return work(txCtx, postgres.ExecutorFor(txCtx, executor))
			})
		}
	} else {
		repository.withinTx = func(ctx context.Context, work func(context.Context, postgres.Executor) error) error {
			return work(ctx, executor)
		}
	}

	return repository
}

// Cast atomically inserts a vote and increments the room score once.
func (repository *Repository) Cast(ctx context.Context, roomID int64, playerID int64) (roomvotes.Mutation, error) {
	result := roomvotes.Mutation{}
	err := repository.withinTx(ctx, func(txCtx context.Context, executor postgres.Executor) error {
		tag, err := executor.Exec(txCtx, insertVoteSQL, roomID, playerID)
		if err != nil {
			return fmt.Errorf("insert room vote: %w", err)
		}
		result.Inserted = tag.RowsAffected() == 1
		query := selectScoreSQL
		if result.Inserted {
			query = incrementScoreSQL
		}
		if err := executor.QueryRow(txCtx, query, roomID).Scan(&result.Score); err != nil {
			return fmt.Errorf("read resulting room score: %w", err)
		}

		return nil
	})

	return result, err
}

// HasVote reports whether one player voted for one room.
func (repository *Repository) HasVote(ctx context.Context, roomID int64, playerID int64) (bool, error) {
	var voted bool
	if err := repository.executor.QueryRow(ctx, hasVoteSQL, roomID, playerID).Scan(&voted); err != nil {
		return false, fmt.Errorf("query room vote: %w", err)
	}

	return voted, nil
}

// Existing returns voters present in a supplied player id set.
func (repository *Repository) Existing(ctx context.Context, roomID int64, playerIDs []int64) (map[int64]struct{}, error) {
	voters := make(map[int64]struct{}, len(playerIDs))
	if len(playerIDs) == 0 {
		return voters, nil
	}
	rows, err := repository.executor.Query(ctx, existingVotesSQL, roomID, playerIDs)
	if err != nil {
		return nil, fmt.Errorf("query active room voters: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var playerID int64
		if err := rows.Scan(&playerID); err != nil {
			return nil, fmt.Errorf("scan active room voter: %w", err)
		}
		voters[playerID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active room voters: %w", err)
	}

	return voters, nil
}

// List returns durable votes matching a query.
func (repository *Repository) List(ctx context.Context, query roomvotes.Query) ([]roomvotes.Vote, error) {
	rows, err := repository.executor.Query(ctx, listVotesSQL, query.RoomID, query.PlayerID, query.Before, query.Limit)
	if err != nil {
		return nil, fmt.Errorf("list room votes: %w", err)
	}
	defer rows.Close()
	votes := make([]roomvotes.Vote, 0, query.Limit)
	for rows.Next() {
		var vote roomvotes.Vote
		if err := rows.Scan(&vote.RoomID, &vote.PlayerID, &vote.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan room vote: %w", err)
		}
		votes = append(votes, vote)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room votes: %w", err)
	}

	return votes, nil
}
