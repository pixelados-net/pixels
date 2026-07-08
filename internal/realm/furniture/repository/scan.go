package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// scanDefinition scans a furniture definition row.
func scanDefinition(row pgx.Row) (furnituremodel.Definition, error) {
	var definition furnituremodel.Definition
	var kind string
	var metadata []byte
	var deletedAt pgtype.Timestamptz
	err := row.Scan(
		&definition.ID,
		&definition.SpriteID,
		&definition.Name,
		&definition.PublicName,
		&kind,
		&definition.Width,
		&definition.Length,
		&definition.StackHeight,
		&definition.AllowStack,
		&definition.AllowWalk,
		&definition.AllowSit,
		&definition.AllowLay,
		&definition.AllowInventoryStack,
		&definition.InteractionType,
		&definition.InteractionModesCount,
		&definition.Multiheight,
		&definition.CustomParams,
		&metadata,
		&definition.CreatedAt,
		&definition.UpdatedAt,
		&deletedAt,
		&definition.Version.Version,
	)
	if err != nil {
		return furnituremodel.Definition{}, err
	}
	definition.Kind = furnituremodel.Kind(kind)
	definition.Metadata = json.RawMessage(metadata)
	definition.DeletedAt = timePointer(deletedAt)

	return definition, nil
}

// scanDefinitions scans furniture definition rows.
func scanDefinitions(rows pgx.Rows) ([]furnituremodel.Definition, error) {
	var definitions []furnituremodel.Definition
	for rows.Next() {
		definition, err := scanDefinition(rows)
		if err != nil {
			return nil, fmt.Errorf("scan furniture definition: %w", err)
		}
		definitions = append(definitions, definition)
	}

	return definitions, rows.Err()
}

// scanItem scans a furniture item row.
func scanItem(row pgx.Row) (furnituremodel.Item, error) {
	var item furnituremodel.Item
	var roomID pgtype.Int8
	var x pgtype.Int2
	var y pgtype.Int2
	var z pgtype.Float8
	var rotation int16
	var wallPosition pgtype.Text
	var metadata []byte
	var deletedAt pgtype.Timestamptz
	err := row.Scan(
		&item.ID,
		&item.DefinitionID,
		&item.OwnerPlayerID,
		&roomID,
		&x,
		&y,
		&z,
		&rotation,
		&wallPosition,
		&item.ExtraData,
		&metadata,
		&item.CreatedAt,
		&item.UpdatedAt,
		&deletedAt,
		&item.Version.Version,
	)
	if err != nil {
		return furnituremodel.Item{}, err
	}
	item.RoomID = int64Pointer(roomID)
	item.X = int2Pointer(x)
	item.Y = int2Pointer(y)
	item.Z = float64Pointer(z)
	item.Rotation = furnituremodel.Rotation(rotation)
	item.WallPosition = stringPointer(wallPosition)
	item.Metadata = json.RawMessage(metadata)
	item.DeletedAt = timePointer(deletedAt)

	return item, nil
}

// scanItems scans furniture item rows.
func scanItems(rows pgx.Rows) ([]furnituremodel.Item, error) {
	var items []furnituremodel.Item
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan furniture item: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// timePointer converts a PostgreSQL timestamp to an optional time.
func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}

// int64Pointer converts a PostgreSQL int8 to an optional int64.
func int64Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}

	return &value.Int64
}

// int2Pointer converts a PostgreSQL int2 to an optional int.
func int2Pointer(value pgtype.Int2) *int {
	if !value.Valid {
		return nil
	}
	result := int(value.Int16)

	return &result
}

// float64Pointer converts a PostgreSQL float8 to an optional float64.
func float64Pointer(value pgtype.Float8) *float64 {
	if !value.Valid {
		return nil
	}

	return &value.Float64
}

// stringPointer converts a PostgreSQL text value to an optional string.
func stringPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}
