// Package promotion contains PostgreSQL room-promotion persistence.
package promotion

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	roompromotion "github.com/niflaot/pixels/internal/realm/room/promotion"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	upsertSQL = `INSERT INTO room_promotions (room_id, category_id, title, description, starts_at, ends_at, created_by)
VALUES ($1,$2,$3,$4,$7,$7 + ($8 * interval '1 second'),$5)
ON CONFLICT (room_id) DO UPDATE SET category_id=EXCLUDED.category_id,title=EXCLUDED.title,description=EXCLUDED.description,
starts_at=CASE WHEN $6 AND room_promotions.ends_at>$7 THEN room_promotions.starts_at ELSE $7 END,
ends_at=CASE WHEN $6 AND room_promotions.ends_at>$7 THEN room_promotions.ends_at + ($8 * interval '1 second') ELSE $7 + ($8 * interval '1 second') END,
created_by=EXCLUDED.created_by,updated_at=$7,version=room_promotions.version+1
RETURNING id,room_id,category_id,title,description,starts_at,ends_at,created_by,version`
	findActiveSQL   = `SELECT id,room_id,category_id,title,description,starts_at,ends_at,created_by,version FROM room_promotions WHERE room_id=$1 AND ends_at>NOW()`
	findByIDSQL     = `SELECT id,room_id,category_id,title,description,starts_at,ends_at,created_by,version FROM room_promotions WHERE id=$1`
	updateCopySQL   = `UPDATE room_promotions SET title=$3,description=$4,updated_at=NOW(),version=version+1 WHERE id=$1 AND created_by=$2 AND ends_at>NOW() RETURNING id,room_id,category_id,title,description,starts_at,ends_at,created_by,version`
	activeRoomsSQL  = `SELECT room_id FROM room_promotions WHERE room_id=ANY($1::bigint[]) AND ends_at>NOW()`
	deleteByRoomSQL = `DELETE FROM room_promotions WHERE room_id=$1`
)

// Repository persists room promotions.
type Repository struct {
	executor postgres.Executor
	pool     *postgres.Pool
}

// DeleteByRoom force-cancels one room promotion.
func (repository *Repository) DeleteByRoom(ctx context.Context, roomID int64) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, deleteByRoomSQL, roomID)
	if err != nil {
		return false, fmt.Errorf("delete room promotion: %w", err)
	}
	return result.RowsAffected() == 1, nil
}

// New creates PostgreSQL room-promotion persistence.
func New(pool *postgres.Pool) *Repository { return &Repository{executor: pool, pool: pool} }

// WithinTransaction runs work inside the shared transaction scope.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, ok := postgres.ScopedExecutor(ctx); ok {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

// Upsert creates, replaces, or extends one room promotion atomically.
func (repository *Repository) Upsert(ctx context.Context, params roompromotion.PurchaseParams, config roompromotion.Config) (roompromotion.Promotion, error) {
	row := repository.executorFor(ctx).QueryRow(ctx, upsertSQL, params.RoomID, params.CategoryID, params.Title, params.Description, params.PlayerID, params.Extended, time.Now().UTC(), int64(config.Duration/time.Second))
	value, err := scan(row)
	if err != nil {
		return roompromotion.Promotion{}, fmt.Errorf("upsert room promotion: %w", err)
	}
	return value, nil
}

// FindActiveByRoom finds one active room promotion.
func (repository *Repository) FindActiveByRoom(ctx context.Context, roomID int64) (roompromotion.Promotion, bool, error) {
	value, err := scan(repository.executorFor(ctx).QueryRow(ctx, findActiveSQL, roomID))
	if err == pgx.ErrNoRows {
		return roompromotion.Promotion{}, false, nil
	}
	if err != nil {
		return roompromotion.Promotion{}, false, fmt.Errorf("find active room promotion: %w", err)
	}
	return value, true, nil
}

// FindByID finds one room promotion.
func (repository *Repository) FindByID(ctx context.Context, id int64) (roompromotion.Promotion, bool, error) {
	value, err := scan(repository.executorFor(ctx).QueryRow(ctx, findByIDSQL, id))
	if err == pgx.ErrNoRows {
		return roompromotion.Promotion{}, false, nil
	}
	if err != nil {
		return roompromotion.Promotion{}, false, fmt.Errorf("find room promotion: %w", err)
	}
	return value, true, nil
}

// UpdateCopy changes active promotion copy for its creator.
func (repository *Repository) UpdateCopy(ctx context.Context, params roompromotion.EditParams) (roompromotion.Promotion, bool, error) {
	value, err := scan(repository.executorFor(ctx).QueryRow(ctx, updateCopySQL, params.PromotionID, params.PlayerID, params.Title, params.Description))
	if err == pgx.ErrNoRows {
		return roompromotion.Promotion{}, false, nil
	}
	if err != nil {
		return roompromotion.Promotion{}, false, fmt.Errorf("update room promotion: %w", err)
	}
	return value, true, nil
}

// ActiveRoomIDs returns promoted ids from a bounded room set.
func (repository *Repository) ActiveRoomIDs(ctx context.Context, roomIDs []int64) (map[int64]struct{}, error) {
	result := make(map[int64]struct{}, len(roomIDs))
	if len(roomIDs) == 0 {
		return result, nil
	}
	rows, err := repository.executorFor(ctx).Query(ctx, activeRoomsSQL, roomIDs)
	if err != nil {
		return nil, fmt.Errorf("list active room promotions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var roomID int64
		if err = rows.Scan(&roomID); err != nil {
			return nil, fmt.Errorf("scan active room promotion: %w", err)
		}
		result[roomID] = struct{}{}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active room promotions: %w", err)
	}
	return result, nil
}

func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}
func scan(row pgx.Row) (value roompromotion.Promotion, err error) {
	err = row.Scan(&value.ID, &value.RoomID, &value.CategoryID, &value.Title, &value.Description, &value.StartsAt, &value.EndsAt, &value.CreatedBy, &value.Version)
	return value, err
}

var _ roompromotion.Store = (*Repository)(nil)
