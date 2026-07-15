package database

import (
	"context"

	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
)

// Issues lists bounded moderation issues by state.
func (repository *Repository) Issues(ctx context.Context, state string, limit int32) ([]moderationrecord.Issue, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := repository.executor(ctx).Query(ctx, issueSelect+` where ($1='' or state=$1) order by created_at limit $2`, state, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]moderationrecord.Issue, 0)
	for rows.Next() {
		value, scanErr := scanIssue(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// Pending lists unresolved issues owned by one reporter.
func (repository *Repository) Pending(ctx context.Context, reporterID int64) ([]moderationrecord.Issue, error) {
	rows, err := repository.executor(ctx).Query(ctx, issueSelect+` where reporter_player_id=$1 and state in ('open','picked') order by created_at desc`, reporterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]moderationrecord.Issue, 0)
	for rows.Next() {
		value, scanErr := scanIssue(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// DeletePending removes only the reporter's unresolved issues.
func (repository *Repository) DeletePending(ctx context.Context, reporterID int64) ([]int64, error) {
	rows, err := repository.executor(ctx).Query(ctx, `update moderation_issues set state='deleted',closed_at=now() where reporter_player_id=$1 and state in ('open','picked') returning id`, reporterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Pick atomically claims one open issue.
func (repository *Repository) Pick(ctx context.Context, id int64, moderatorID int64) (moderationrecord.Issue, bool, error) {
	return repository.mutateIssue(ctx, `update moderation_issues set state='picked',picked_by_player_id=$2,picked_at=now() where id=$1 and state='open' returning id,reporter_player_id,reported_player_id,room_id,topic_id,kind,message,state,resolution,picked_by_player_id,picked_at,closed_by_player_id,closed_at,created_at`, id, moderatorID)
}

// Release returns one moderator-owned issue to the queue.
func (repository *Repository) Release(ctx context.Context, id int64, moderatorID int64) (moderationrecord.Issue, bool, error) {
	return repository.mutateIssue(ctx, `update moderation_issues set state='open',picked_by_player_id=null,picked_at=null where id=$1 and state='picked' and picked_by_player_id=$2 returning id,reporter_player_id,reported_player_id,room_id,topic_id,kind,message,state,resolution,picked_by_player_id,picked_at,closed_by_player_id,closed_at,created_at`, id, moderatorID)
}

// Close resolves one assigned issue.
func (repository *Repository) Close(ctx context.Context, id int64, moderatorID int64, resolution int32) (moderationrecord.Issue, bool, error) {
	rows, err := repository.executor(ctx).Query(ctx, `update moderation_issues set state='resolved',resolution=$3,closed_by_player_id=$2,closed_at=now() where id=$1 and state in ('open','picked') and (picked_by_player_id is null or picked_by_player_id=$2) returning id,reporter_player_id,reported_player_id,room_id,topic_id,kind,message,state,resolution,picked_by_player_id,picked_at,closed_by_player_id,closed_at,created_at`, id, moderatorID, resolution)
	return scanMutation(rows, err)
}

// mutateIssue runs one two-argument conditional issue mutation.
func (repository *Repository) mutateIssue(ctx context.Context, statement string, id int64, moderatorID int64) (moderationrecord.Issue, bool, error) {
	rows, err := repository.executor(ctx).Query(ctx, statement, id, moderatorID)
	return scanMutation(rows, err)
}

// scanMutation maps an optional mutation result.
func scanMutation(rows interface {
	Next() bool
	Scan(...any) error
	Close()
}, err error) (moderationrecord.Issue, bool, error) {
	if err != nil {
		return moderationrecord.Issue{}, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return moderationrecord.Issue{}, false, nil
	}
	value, err := scanIssue(rows)
	return value, err == nil, err
}
