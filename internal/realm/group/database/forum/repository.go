// Package forum implements social-group forum persistence in PostgreSQL.
package forum

import (
	"context"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository stores forum threads, posts, and read markers.
type Repository struct {
	// pool executes PostgreSQL operations.
	pool *postgres.Pool
}

// New creates a forum repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// executor returns the active scoped transaction.
func (repository *Repository) executor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.pool)
}

// WithinTransaction runs work atomically.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, scoped := postgres.ScopedExecutor(ctx); scoped {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

// threadScanner scans one durable thread row.
type threadScanner interface{ Scan(...any) error }

// scanThread maps one durable thread row.
func scanThread(row threadScanner) (grouprecord.Thread, error) {
	var thread grouprecord.Thread
	err := row.Scan(&thread.ID, &thread.GroupID, &thread.AuthorID, &thread.AuthorName, &thread.Subject, &thread.State,
		&thread.Pinned, &thread.Locked, &thread.PostCount, &thread.UnreadCount, &thread.LastPostID, &thread.LastAuthorID, &thread.LastAuthorName,
		&thread.LastPostedAt, &thread.ModeratorID, &thread.ModeratorName, &thread.ModerationReason, &thread.ModeratedAt,
		&thread.CreatedAt, &thread.UpdatedAt, &thread.Version)
	return thread, err
}

// postScanner scans one durable post row.
type postScanner interface{ Scan(...any) error }

// scanPost maps one durable post row.
func scanPost(row postScanner) (grouprecord.Post, error) {
	var post grouprecord.Post
	err := row.Scan(&post.ID, &post.GroupID, &post.ThreadID, &post.Ordinal, &post.AuthorID, &post.AuthorName,
		&post.AuthorFigure, &post.Body, &post.State, &post.ModeratorID, &post.ModeratorName, &post.ModerationReason,
		&post.ModeratedAt, &post.AuthorPostCount, &post.CreatedAt, &post.UpdatedAt, &post.Version)
	return post, err
}
