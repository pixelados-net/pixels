package database

import (
	"context"
	"time"

	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
)

// ClaimCleanup claims stale unreferenced objects without holding locks during storage I/O.
func (repository *Repository) ClaimCleanup(ctx context.Context, pendingBefore time.Time, supersededBefore time.Time, retryBefore time.Time, limit int) ([]camerarecord.CleanupCandidate, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `with candidates as (
        select capture.id from camera_captures capture
        where capture.kind='photo'
          and ((capture.state='pending' and capture.created_at<$1)
            or (capture.state='superseded' and capture.superseded_at<$2)
            or (capture.state='abandoned' and capture.cleanup_attempted_at<$3))
          and not exists(select 1 from camera_capture_items link where link.capture_id=capture.id)
          and not exists(select 1 from camera_publications publication where publication.capture_id=capture.id and publication.removed_at is null)
        order by capture.created_at,capture.id
        for update skip locked limit $4
    )
    update camera_captures capture set
        state='abandoned',abandoned_at=coalesce(abandoned_at,now()),
        cleanup_attempted_at=now(),consumed_at=coalesce(consumed_at,now()),version=version+1
    from candidates where capture.id=candidates.id
    returning capture.id,capture.storage_key`, pendingBefore, supersededBefore, retryBefore, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]camerarecord.CleanupCandidate, 0, limit)
	for rows.Next() {
		var value camerarecord.CleanupCandidate
		if err = rows.Scan(&value.CaptureID, &value.StorageKey); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// MarkDeleted records one successful object deletion idempotently.
func (repository *Repository) MarkDeleted(ctx context.Context, captureID int64, deletedAt time.Time) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update camera_captures set state='deleted',deleted_at=$2,version=version+1 where id=$1 and state='abandoned'`, captureID, deletedAt)
	return err == nil && result.RowsAffected() == 1, err
}
