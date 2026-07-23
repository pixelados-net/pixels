package repository

import (
	"time"

	"github.com/jackc/pgx/v5"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
)

// scanCustomLayout scans one room-owned custom layout row.
func scanCustomLayout(row pgx.Row) (roomlayout.Layout, error) {
	var roomLayout roomlayout.Layout
	var updatedAt time.Time
	err := row.Scan(
		&roomLayout.RoomID, &roomLayout.Name, &roomLayout.Heightmap,
		&roomLayout.DoorX, &roomLayout.DoorY, &roomLayout.DoorDirection,
		&roomLayout.WallThickness, &roomLayout.FloorThickness,
		&roomLayout.WallHeight, &updatedAt,
	)
	roomLayout.UpdatedAt = updatedAt

	return roomLayout, err
}
