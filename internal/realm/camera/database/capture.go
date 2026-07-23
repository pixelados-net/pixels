package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
)

// captureColumns stores the canonical camera capture projection.
const captureColumns = `id, capture_uuid::text, player_id, room_id, kind, state, storage_key, url, created_at, consumed_at, superseded_at, abandoned_at, deleted_at, cleanup_attempted_at, purchase_count, version`

// CreateCapture inserts one uploaded camera artifact.
func (repository *Repository) CreateCapture(ctx context.Context, capture camerarecord.Capture) (camerarecord.Capture, error) {
	var created camerarecord.Capture
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		if capture.Kind == camerarecord.KindPhoto {
			if _, err := repository.executorFor(txCtx).Exec(txCtx, `update camera_captures set state='superseded',superseded_at=now(),consumed_at=coalesce(consumed_at,now()),version=version+1 where player_id=$1 and kind='photo' and state in ('pending','purchased','published','purchased_published')`, capture.PlayerID); err != nil {
				return err
			}
		}
		row := repository.executorFor(txCtx).QueryRow(txCtx, `insert into camera_captures (capture_uuid, player_id, room_id, kind, storage_key, url, consumed_at) values ($1::uuid,$2,$3,$4,$5,$6,$7) returning `+captureColumns, capture.UUID, capture.PlayerID, capture.RoomID, capture.Kind, capture.StorageKey, capture.URL, capture.ConsumedAt)
		var err error
		created, err = scanCapture(row)
		return err
	})
	return created, err
}

// ActiveCapture locks and returns the current reusable player photo.
func (repository *Repository) ActiveCapture(ctx context.Context, playerID int64) (camerarecord.Capture, bool, error) {
	row := repository.executorFor(ctx).QueryRow(ctx, `select `+captureColumns+` from camera_captures where player_id=$1 and kind='photo' and state in ('pending','purchased','published','purchased_published') order by created_at desc,id desc limit 1 for update`, playerID)
	capture, err := scanCapture(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return camerarecord.Capture{}, false, nil
	}
	return capture, err == nil, err
}

// AttachPurchase links one furniture copy and advances active capture state.
func (repository *Repository) AttachPurchase(ctx context.Context, captureID int64, itemID int64) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `with linked as (
        insert into camera_capture_items(capture_id,item_id) values($1,$2)
        on conflict do nothing returning 1
    )
    update camera_captures set
        purchase_count=purchase_count+(select count(*) from linked),
        state=case when state='pending' then 'purchased' when state='published' then 'purchased_published' else state end,
        version=version+(select count(*) from linked)
    where id=$1`, captureID, itemID)
	return err
}

// LatestCaptureAt returns the latest capture time for one player and kind.
func (repository *Repository) LatestCaptureAt(ctx context.Context, playerID int64, kind camerarecord.Kind) (time.Time, bool, error) {
	var createdAt time.Time
	err := repository.executorFor(ctx).QueryRow(ctx, `select created_at from camera_captures where player_id=$1 and kind=$2 order by created_at desc, id desc limit 1`, playerID, kind).Scan(&createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}
	return createdAt, err == nil, err
}

// Captures lists recent captures for one player.
func (repository *Repository) Captures(ctx context.Context, playerID int64, limit int) ([]camerarecord.Capture, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select `+captureColumns+` from camera_captures where player_id=$1 order by created_at desc, id desc limit $2`, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]camerarecord.Capture, 0, limit)
	for rows.Next() {
		capture, scanErr := scanCapture(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, capture)
	}
	return result, rows.Err()
}

// captureScanner scans one camera capture row.
type captureScanner interface {
	// Scan decodes one database row.
	Scan(...any) error
}

// scanCapture scans one camera capture row.
func scanCapture(scanner captureScanner) (camerarecord.Capture, error) {
	var capture camerarecord.Capture
	err := scanner.Scan(&capture.ID, &capture.UUID, &capture.PlayerID, &capture.RoomID, &capture.Kind, &capture.State, &capture.StorageKey, &capture.URL, &capture.CreatedAt, &capture.ConsumedAt, &capture.SupersededAt, &capture.AbandonedAt, &capture.DeletedAt, &capture.CleanupAttemptedAt, &capture.PurchaseCount, &capture.Version)
	return capture, err
}
