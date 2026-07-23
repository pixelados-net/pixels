package database

import (
	"context"
	"time"

	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// CreateIssue atomically inserts one issue and frozen evidence.
func (repository *Repository) CreateIssue(ctx context.Context, params moderationrecord.ReportParams) (moderationrecord.Issue, error) {
	var created moderationrecord.Issue
	err := postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		row := repository.executor(txCtx).QueryRow(txCtx, `insert into moderation_issues(reporter_player_id,reported_player_id,room_id,photo_item_id,topic_id,kind,message) values($1,$2,$3,$4,$5,$6,$7) returning id,reporter_player_id,reported_player_id,room_id,photo_item_id,topic_id,kind,message,state,resolution,picked_by_player_id,picked_at,closed_by_player_id,closed_at,created_at`, params.ReporterPlayerID, params.ReportedPlayerID, params.RoomID, params.PhotoItemID, params.TopicID, params.Kind, params.Message)
		value, scanErr := scanIssue(row)
		if scanErr != nil {
			return scanErr
		}
		for _, entry := range params.Chatlog {
			if _, insertErr := repository.executor(txCtx).Exec(txCtx, `insert into issue_chatlog(issue_id,player_id,pattern_id,message,created_at) values($1,$2,$3,$4,coalesce($5::timestamptz,now()))`, value.ID, entry.PlayerID, entry.PatternID, entry.Message, nullableTime(entry.CreatedAt)); insertErr != nil {
				return insertErr
			}
		}
		value.Chatlog = append([]moderationrecord.ChatEntry(nil), params.Chatlog...)
		created = value
		return nil
	})
	return created, err
}

// Issue returns one issue with optional frozen evidence.
func (repository *Repository) Issue(ctx context.Context, id int64, withChatlog bool) (moderationrecord.Issue, bool, error) {
	rows, err := repository.executor(ctx).Query(ctx, staffIssueSelect+` where i.id=$1`, id)
	if err != nil {
		return moderationrecord.Issue{}, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return moderationrecord.Issue{}, false, rows.Err()
	}
	value, err := scanStaffIssue(rows)
	if err != nil || !withChatlog {
		return value, err == nil, err
	}
	value.Chatlog, err = repository.chatlog(ctx, id)
	return value, err == nil, err
}

// chatlog returns frozen issue evidence.
func (repository *Repository) chatlog(ctx context.Context, id int64) ([]moderationrecord.ChatEntry, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select id,player_id,pattern_id,message,created_at from issue_chatlog where issue_id=$1 order by id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]moderationrecord.ChatEntry, 0)
	for rows.Next() {
		var value moderationrecord.ChatEntry
		if err = rows.Scan(&value.ID, &value.PlayerID, &value.PatternID, &value.Message, &value.CreatedAt); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// nullableTime returns nil for a missing captured timestamp.
func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
