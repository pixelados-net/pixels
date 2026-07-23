// Package database implements PostgreSQL moderation persistence.
package database

import (
	"context"

	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository stores issues, topics, preferences, and guide feedback.
type Repository struct {
	// pool owns database connections.
	pool *postgres.Pool
}

// New creates a moderation repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// executor returns the active transaction or pool.
func (repository *Repository) executor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.pool)
}

// rowScanner scans one database row.
type rowScanner interface{ Scan(...any) error }

// issueSelect reads one persistence-only issue projection.
const issueSelect = `select id,reporter_player_id,reported_player_id,room_id,photo_item_id,topic_id,kind,message,state,resolution,picked_by_player_id,picked_at,closed_by_player_id,closed_at,created_at from moderation_issues`

// staffIssueSelect reads one issue with every Nitro-visible player name in one query.
const staffIssueSelect = `select i.id,i.reporter_player_id,i.reported_player_id,i.room_id,i.photo_item_id,i.topic_id,i.kind,i.message,i.state,i.resolution,i.picked_by_player_id,i.picked_at,i.closed_by_player_id,i.closed_at,i.created_at,reporter.username,coalesce(reported.username,''),coalesce(picker.username,'') from moderation_issues i join players reporter on reporter.id=i.reporter_player_id left join players reported on reported.id=i.reported_player_id left join players picker on picker.id=i.picked_by_player_id`

// scanIssue maps one complete issue row.
func scanIssue(row rowScanner) (moderationrecord.Issue, error) {
	var value moderationrecord.Issue
	err := row.Scan(&value.ID, &value.ReporterPlayerID, &value.ReportedPlayerID, &value.RoomID, &value.PhotoItemID, &value.TopicID, &value.Kind, &value.Message, &value.State, &value.Resolution, &value.PickedByPlayerID, &value.PickedAt, &value.ClosedByPlayerID, &value.ClosedAt, &value.CreatedAt)
	return value, err
}

// scanStaffIssue maps one complete issue row with player display names.
func scanStaffIssue(row rowScanner) (moderationrecord.Issue, error) {
	var value moderationrecord.Issue
	err := row.Scan(&value.ID, &value.ReporterPlayerID, &value.ReportedPlayerID, &value.RoomID, &value.PhotoItemID, &value.TopicID, &value.Kind, &value.Message, &value.State, &value.Resolution, &value.PickedByPlayerID, &value.PickedAt, &value.ClosedByPlayerID, &value.ClosedAt, &value.CreatedAt, &value.ReporterName, &value.ReportedName, &value.PickerName)
	return value, err
}
