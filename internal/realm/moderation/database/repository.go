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

const issueSelect = `select id,reporter_player_id,reported_player_id,room_id,topic_id,kind,message,state,resolution,picked_by_player_id,picked_at,closed_by_player_id,closed_at,created_at from moderation_issues`

// scanIssue maps one complete issue row.
func scanIssue(row rowScanner) (moderationrecord.Issue, error) {
	var value moderationrecord.Issue
	err := row.Scan(&value.ID, &value.ReporterPlayerID, &value.ReportedPlayerID, &value.RoomID, &value.TopicID, &value.Kind, &value.Message, &value.State, &value.Resolution, &value.PickedByPlayerID, &value.PickedAt, &value.ClosedByPlayerID, &value.ClosedAt, &value.CreatedAt)
	return value, err
}
