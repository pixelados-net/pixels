package layout

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// scanLayout scans one room layout row.
func scanLayout(row pgx.Row) (Layout, error) {
	var roomLayout Layout
	var deletedAt pgtype.Timestamptz

	err := row.Scan(
		&roomLayout.ID,
		&roomLayout.Name,
		&roomLayout.TileSize,
		&roomLayout.Heightmap,
		&roomLayout.DoorX,
		&roomLayout.DoorY,
		&roomLayout.DoorZ,
		&roomLayout.DoorDirection,
		&roomLayout.ClubLevel,
		&roomLayout.Enabled,
		&roomLayout.CreatedAt,
		&roomLayout.UpdatedAt,
		&deletedAt,
		&roomLayout.Version.Version,
	)
	if err != nil {
		return Layout{}, err
	}

	roomLayout.DeletedAt = timePointer(deletedAt)

	return roomLayout, nil
}

// scanLayouts scans room layout rows.
func scanLayouts(rows pgx.Rows) ([]Layout, error) {
	var layouts []Layout
	for rows.Next() {
		roomLayout, err := scanLayout(rows)
		if err != nil {
			return nil, fmt.Errorf("scan room layout: %w", err)
		}
		layouts = append(layouts, roomLayout)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read room layouts: %w", err)
	}

	return layouts, nil
}

// timePointer converts a PostgreSQL timestamp to an optional time.
func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}
