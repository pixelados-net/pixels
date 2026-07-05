package repository

import (
	"context"
	"fmt"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/model"
)

const (
	// liftedRoomColumns contains the shared lifted room select list.
	liftedRoomColumns = `id, room_id, area_id, image, caption, order_num, starts_at, ends_at, created_at, updated_at, deleted_at, version`

	// listLiftedRoomsSQL reads active lifted rooms.
	listLiftedRoomsSQL = `
select ` + liftedRoomColumns + `
from navigator_lifted_rooms
where deleted_at is null and (starts_at is null or starts_at <= now()) and (ends_at is null or ends_at > now())
order by order_num asc, id asc`
)

// ListLiftedRooms lists currently active lifted rooms.
func (repository *Repository) ListLiftedRooms(ctx context.Context) ([]navmodel.LiftedRoom, error) {
	rows, err := repository.executor.Query(ctx, listLiftedRoomsSQL)
	if err != nil {
		return nil, fmt.Errorf("list lifted rooms: %w", err)
	}
	defer rows.Close()

	return scanLiftedRooms(rows)
}
