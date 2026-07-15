package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
)

// Preferences returns one moderator's geometry or protocol defaults.
func (repository *Repository) Preferences(ctx context.Context, playerID int64) (moderationrecord.Preferences, error) {
	value := moderationrecord.Preferences{PlayerID: playerID, Width: 640, Height: 480}
	err := repository.executor(ctx).QueryRow(ctx, `select window_x,window_y,window_width,window_height from moderator_preferences where player_id=$1`, playerID).Scan(&value.X, &value.Y, &value.Width, &value.Height)
	if err == pgx.ErrNoRows {
		return value, nil
	}
	return value, err
}

// Visits returns recent rooms entered by one player.
func (repository *Repository) Visits(ctx context.Context, playerID int64, limit int32) ([]moderationrecord.RoomVisit, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := repository.executor(ctx).Query(ctx, `select v.room_id,r.name,v.entered_at from room_visits v join rooms r on r.id=v.room_id where v.player_id=$1 order by v.entered_at desc limit $2`, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]moderationrecord.RoomVisit, 0)
	for rows.Next() {
		var value moderationrecord.RoomVisit
		if err = rows.Scan(&value.RoomID, &value.RoomName, &value.EnteredAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// InsertFeedback stores one guide recommendation.
func (repository *Repository) InsertFeedback(ctx context.Context, guideID int64, requesterID int64, recommended bool) error {
	_, err := repository.executor(ctx).Exec(ctx, `insert into guide_feedback(guide_player_id,requester_player_id,recommended) values($1,$2,$3)`, guideID, requesterID, recommended)
	return err
}

// CreateGuardianTicket stores one peer-review ticket.
func (repository *Repository) CreateGuardianTicket(ctx context.Context, reporterID int64, reportedID int64, closesAt time.Time) (int64, error) {
	var id int64
	err := repository.executor(ctx).QueryRow(ctx, `insert into guardian_tickets(reporter_player_id,reported_player_id,closes_at) values($1,$2,$3) returning id`, reporterID, reportedID, closesAt).Scan(&id)
	return id, err
}

// SaveGuardianVote records one guardian verdict idempotently.
func (repository *Repository) SaveGuardianVote(ctx context.Context, ticketID int64, guardianID int64, vote int32) error {
	_, err := repository.executor(ctx).Exec(ctx, `insert into guardian_votes(ticket_id,guardian_player_id,vote) values($1,$2,$3) on conflict(ticket_id,guardian_player_id) do update set vote=excluded.vote,created_at=now()`, ticketID, guardianID, vote)
	return err
}

// CloseGuardianTicket stores one final peer-review result.
func (repository *Repository) CloseGuardianTicket(ctx context.Context, ticketID int64, result int32) error {
	_, err := repository.executor(ctx).Exec(ctx, `update guardian_tickets set state='closed',result=$2 where id=$1`, ticketID, result)
	return err
}

var _ moderationrecord.Store = (*Repository)(nil)
