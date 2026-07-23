package database

import (
	"context"
	"fmt"
	"time"

	navrecord "github.com/niflaot/pixels/internal/realm/navigator/record"
)

// RecordVisit upserts one admitted room visit.
func (repository *Repository) RecordVisit(ctx context.Context, playerID int64, roomID int64) error {
	_, err := repository.executor.Exec(ctx, `insert into navigator_room_visits(player_id,room_id) values($1,$2) on conflict(player_id,room_id) do update set visit_count=navigator_room_visits.visit_count+1,last_visited_at=now()`, playerID, roomID)
	if err != nil {
		return fmt.Errorf("record navigator room visit: %w", err)
	}
	return nil
}

// RecordVisits upserts one coalesced batch in a single database round trip.
func (repository *Repository) RecordVisits(ctx context.Context, visits []navrecord.Visit) error {
	playerIDs := make([]int64, len(visits))
	roomIDs := make([]int64, len(visits))
	visitedAt := make([]time.Time, len(visits))
	increments := make([]bool, len(visits))
	for index, visit := range visits {
		playerIDs[index], roomIDs[index], visitedAt[index], increments[index] = visit.PlayerID, visit.RoomID, visit.VisitedAt, visit.Increment
	}
	_, err := repository.executor.Exec(ctx, `with input(player_id,room_id,visited_at,increment) as (select * from unnest($1::bigint[],$2::bigint[],$3::timestamptz[],$4::boolean[])) insert into navigator_room_visits as visits(player_id,room_id,visit_count,first_visited_at,last_visited_at) select player_id,room_id,1,visited_at,visited_at from input on conflict(player_id,room_id) do update set visit_count=visits.visit_count+case when (select increment from input where player_id=excluded.player_id and room_id=excluded.room_id) then 1 else 0 end,last_visited_at=greatest(visits.last_visited_at,excluded.last_visited_at)`, playerIDs, roomIDs, visitedAt, increments)
	if err != nil {
		return fmt.Errorf("record navigator room visits: %w", err)
	}
	return nil
}

// ListRecentRoomIDs lists active room ids by latest visit.
func (repository *Repository) ListRecentRoomIDs(ctx context.Context, playerID int64, limit int) ([]int64, error) {
	return repository.listVisitRoomIDs(ctx, `select v.room_id from navigator_room_visits v join rooms r on r.id=v.room_id where v.player_id=$1 and r.deleted_at is null and not r.is_bundle_template order by v.last_visited_at desc limit $2`, playerID, limit)
}

// ListFrequentRoomIDs lists active room ids by frequency and recency.
func (repository *Repository) ListFrequentRoomIDs(ctx context.Context, playerID int64, limit int) ([]int64, error) {
	return repository.listVisitRoomIDs(ctx, `select v.room_id from navigator_room_visits v join rooms r on r.id=v.room_id where v.player_id=$1 and r.deleted_at is null and not r.is_bundle_template order by v.visit_count desc,v.last_visited_at desc limit $2`, playerID, limit)
}

// DeleteVisitHistory deletes all retained visits for one player.
func (repository *Repository) DeleteVisitHistory(ctx context.Context, playerID int64) (int64, error) {
	tag, err := repository.executor.Exec(ctx, `delete from navigator_room_visits where player_id=$1`, playerID)
	if err != nil {
		return 0, fmt.Errorf("delete navigator room visits: %w", err)
	}
	return tag.RowsAffected(), nil
}

// listVisitRoomIDs scans one bounded visit ordering.
func (repository *Repository) listVisitRoomIDs(ctx context.Context, query string, playerID int64, limit int) ([]int64, error) {
	rows, err := repository.executor.Query(ctx, query, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]int64, 0, limit)
	for rows.Next() {
		var roomID int64
		if err = rows.Scan(&roomID); err != nil {
			return nil, err
		}
		values = append(values, roomID)
	}
	return values, rows.Err()
}
