// Package games implements room game score persistence in PostgreSQL.
package games

import (
	"context"

	roomgames "github.com/niflaot/pixels/internal/realm/room/world/games"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository persists room game results.
type Repository struct {
	// pool starts shared transaction scopes.
	pool *postgres.Pool
}

// New creates a room game score repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// Save inserts one completed match atomically.
func (repository *Repository) Save(ctx context.Context, entries []roomgames.Score) error {
	return postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		for _, entry := range entries {
			if _, err := executor.Exec(txCtx, `insert into room_game_scores(room_id,game_kind,started_at,player_id,team,score,team_score) values($1,$2,$3,$4,$5,$6,$7)`, entry.RoomID, entry.Kind, entry.StartedAt, entry.PlayerID, entry.Team, entry.Score, entry.TeamScore); err != nil {
				return err
			}
		}
		return nil
	})
}

// List returns a cursor-paginated room history.
func (repository *Repository) List(ctx context.Context, roomID int64, beforeID int64, limit int) (roomgames.ScorePage, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	rows, err := repository.pool.Query(ctx, `select id,game_kind,started_at,player_id,team,score,team_score from room_game_scores where room_id=$1 and ($2=0 or id<$2) order by id desc limit $3`, roomID, beforeID, limit+1)
	if err != nil {
		return roomgames.ScorePage{}, err
	}
	defer rows.Close()
	page := roomgames.ScorePage{Entries: make([]roomgames.Score, 0, limit)}
	for rows.Next() {
		var entry roomgames.Score
		entry.RoomID = roomID
		if err = rows.Scan(&entry.ID, &entry.Kind, &entry.StartedAt, &entry.PlayerID, &entry.Team, &entry.Score, &entry.TeamScore); err != nil {
			return roomgames.ScorePage{}, err
		}
		if len(page.Entries) == limit {
			page.NextID = page.Entries[len(page.Entries)-1].ID
			break
		}
		page.Entries = append(page.Entries, entry)
	}
	return page, rows.Err()
}

var scoreStoreAssertion roomgames.ScoreStore = (*Repository)(nil)
