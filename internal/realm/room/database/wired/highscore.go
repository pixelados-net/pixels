package wired

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Reset deletes every period entry for selected highscore boards in one room.
func (repository *Repository) Reset(ctx context.Context, roomID int64, boardIDs []int64) (int64, error) {
	result, err := repository.pool.Exec(ctx, `delete from room_wired_highscore_entries entry using furniture_items item,furniture_definitions definition where entry.board_item_id=item.id and item.definition_id=definition.id and item.room_id=$1 and item.id=any($2) and definition.interaction_type='wf_highscore'`, roomID, boardIDs)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// RecordAndList atomically records game results and returns a bounded ranking.
func (repository *Repository) RecordAndList(ctx context.Context, boardID int64, mode roomwired.HighscoreMode, period roomwired.HighscorePeriod, periodStart *time.Time, results []roomwired.HighscoreResult, limit int) ([]roomwired.HighscoreEntry, error) {
	var entries []roomwired.HighscoreEntry
	err := repository.within(ctx, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		for _, result := range results {
			ids := sortedIDs(result.PlayerIDs)
			if len(ids) == 0 {
				continue
			}
			key := participantKey(ids)
			wins := int64(0)
			if mode == roomwired.HighscoreMostWins && result.Won {
				wins = 1
			}
			var entryID int64
			err := executor.QueryRow(txCtx, `insert into room_wired_highscore_entries(board_item_id,period_kind,period_start,participant_key,score,wins) values($1,$2,$3,$4,$5,$6) on conflict(board_item_id,period_kind,period_start,participant_key) do update set score=greatest(room_wired_highscore_entries.score,excluded.score),wins=room_wired_highscore_entries.wins+excluded.wins,updated_at=now() returning id`, boardID, period, periodStart, key, result.Score, wins).Scan(&entryID)
			if err != nil {
				return err
			}
			if err = repository.replaceParticipants(txCtx, executor, entryID, ids); err != nil {
				return err
			}
		}
		var err error
		entries, err = repository.listHighscores(txCtx, executor, boardID, mode, period, periodStart, limit)
		return err
	})
	return entries, err
}

// replaceParticipants makes normalized participant order authoritative.
func (repository *Repository) replaceParticipants(ctx context.Context, executor postgres.Executor, entryID int64, playerIDs []int64) error {
	if _, err := executor.Exec(ctx, `delete from room_wired_highscore_participants where entry_id=$1`, entryID); err != nil {
		return err
	}
	for ordinal, playerID := range playerIDs {
		if _, err := executor.Exec(ctx, `insert into room_wired_highscore_participants(entry_id,player_id,ordinal) values($1,$2,$3)`, entryID, playerID, ordinal); err != nil {
			return err
		}
	}
	return nil
}

// listHighscores loads ranked rows and current usernames without global state.
func (repository *Repository) listHighscores(ctx context.Context, executor postgres.Executor, boardID int64, mode roomwired.HighscoreMode, period roomwired.HighscorePeriod, periodStart *time.Time, limit int) ([]roomwired.HighscoreEntry, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	order := "entry.score"
	if mode == roomwired.HighscoreMostWins {
		order = "entry.wins"
	}
	rows, err := executor.Query(ctx, `select entry.id,case when $4 then entry.wins else entry.score end from room_wired_highscore_entries entry where entry.board_item_id=$1 and entry.period_kind=$2 and entry.period_start is not distinct from $3 order by `+order+` desc,entry.updated_at asc,entry.id asc limit $5`, boardID, period, periodStart, mode == roomwired.HighscoreMostWins, limit)
	if err != nil {
		return nil, err
	}
	entryIDs := make([]int64, 0, limit)
	scores := make([]int64, 0, limit)
	for rows.Next() {
		var entryID int64
		var score int64
		if err = rows.Scan(&entryID, &score); err != nil {
			rows.Close()
			return nil, err
		}
		entryIDs = append(entryIDs, entryID)
		scores = append(scores, score)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	entries := make([]roomwired.HighscoreEntry, 0, len(entryIDs))
	for index, entryID := range entryIDs {
		entry := roomwired.HighscoreEntry{Score: scores[index]}
		entry.PlayerIDs, entry.Usernames, err = repository.loadParticipants(ctx, executor, entryID)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// loadParticipants loads one ranked composition in stable order.
func (repository *Repository) loadParticipants(ctx context.Context, executor postgres.Executor, entryID int64) ([]int64, []string, error) {
	rows, err := executor.Query(ctx, `select participant.player_id,player.username from room_wired_highscore_participants participant join players player on player.id=participant.player_id where participant.entry_id=$1 order by participant.ordinal`, entryID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var ids []int64
	var names []string
	for rows.Next() {
		var id int64
		var name string
		if err = rows.Scan(&id, &name); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		names = append(names, name)
	}
	return ids, names, rows.Err()
}

// sortedIDs returns a copied stable participant identity.
func sortedIDs(values []int64) []int64 {
	result := append([]int64(nil), values...)
	sort.Slice(result, func(left int, right int) bool { return result[left] < result[right] })
	return result
}

// participantKey creates a collision-free decimal composition key.
func participantKey(values []int64) string {
	var builder strings.Builder
	for index, value := range values {
		if index > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(strconv.FormatInt(value, 10))
	}
	return builder.String()
}
