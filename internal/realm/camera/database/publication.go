package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
)

// publicationColumns stores the canonical camera publication projection.
const publicationColumns = `id, capture_id, player_id, room_id, url, created_at, removed_at, coalesce(removed_reason,'')`

// PublishCooldown returns one player's last publication time.
func (repository *Repository) PublishCooldown(ctx context.Context, playerID int64) (time.Time, bool, error) {
	var publishedAt time.Time
	err := repository.executorFor(ctx).QueryRow(ctx, `select last_published_at from camera_publish_cooldowns where player_id=$1 for update`, playerID).Scan(&publishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}
	return publishedAt, err == nil, err
}

// SetPublishCooldown upserts one player's publication time.
func (repository *Repository) SetPublishCooldown(ctx context.Context, playerID int64, publishedAt time.Time) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into camera_publish_cooldowns (player_id,last_published_at) values ($1,$2) on conflict (player_id) do update set last_published_at=excluded.last_published_at`, playerID, publishedAt)
	return err
}

// CreatePublication inserts one public gallery entry.
func (repository *Repository) CreatePublication(ctx context.Context, capture camerarecord.Capture) (camerarecord.Publication, error) {
	row := repository.executorFor(ctx).QueryRow(ctx, `insert into camera_publications (capture_id,player_id,room_id,url) values ($1,$2,$3,$4) returning `+publicationColumns, capture.ID, capture.PlayerID, capture.RoomID, capture.URL)
	publication, err := scanPublication(row)
	if err != nil {
		return camerarecord.Publication{}, err
	}
	_, err = repository.executorFor(ctx).Exec(ctx, `update camera_captures set state=case when state='pending' then 'published' when state='purchased' then 'purchased_published' else state end,version=version+1 where id=$1`, capture.ID)
	return publication, err
}

// PublicationByCapture returns one existing active or removed publication.
func (repository *Repository) PublicationByCapture(ctx context.Context, captureID int64) (camerarecord.Publication, bool, error) {
	row := repository.executorFor(ctx).QueryRow(ctx, `select `+publicationColumns+` from camera_publications where capture_id=$1`, captureID)
	publication, err := scanPublication(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return camerarecord.Publication{}, false, nil
	}
	return publication, err == nil, err
}

// Publications lists public gallery entries.
func (repository *Repository) Publications(ctx context.Context, limit int, offset int, includeRemoved bool) ([]camerarecord.Publication, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select `+publicationColumns+` from camera_publications where ($1 or removed_at is null) order by created_at desc,id desc limit $2 offset $3`, includeRemoved, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]camerarecord.Publication, 0, limit)
	for rows.Next() {
		publication, scanErr := scanPublication(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, publication)
	}
	return result, rows.Err()
}

// RemovePublication soft-removes one gallery entry idempotently.
func (repository *Repository) RemovePublication(ctx context.Context, publicationID int64, reason string) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update camera_publications set removed_at=now(),removed_reason=$2 where id=$1 and removed_at is null`, publicationID, reason)
	return err == nil && result.RowsAffected() == 1, err
}

// InsertAudit appends one administrative mutation record.
func (repository *Repository) InsertAudit(ctx context.Context, audit camerarecord.Audit) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into camera_audit_log (actor_player_id,action,entity_id,reason) values ($1,$2,nullif($3,0),$4)`, audit.ActorPlayerID, audit.Action, audit.EntityID, audit.Reason)
	return err
}

// publicationScanner scans one camera publication row.
type publicationScanner interface {
	// Scan decodes one database row.
	Scan(...any) error
}

// scanPublication scans one camera publication row.
func scanPublication(scanner publicationScanner) (camerarecord.Publication, error) {
	var publication camerarecord.Publication
	err := scanner.Scan(&publication.ID, &publication.CaptureID, &publication.PlayerID, &publication.RoomID, &publication.URL, &publication.CreatedAt, &publication.RemovedAt, &publication.RemovedReason)
	return publication, err
}
