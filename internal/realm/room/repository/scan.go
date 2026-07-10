package repository

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
)

// scanRoom scans a room row.
func scanRoom(row pgx.Row) (roommodel.Room, error) {
	var room roommodel.Room
	var categoryID pgtype.Int8
	var passwordHash pgtype.Text
	var deletedAt pgtype.Timestamptz
	var doorMode int16
	var tradeMode int16
	var moderationMute int16
	var moderationKick int16
	var moderationBan int16
	err := row.Scan(&room.ID, &room.OwnerPlayerID, &room.OwnerName, &room.Name, &room.Description, &room.ModelName, &doorMode, &passwordHash, &room.MaxUsers, &room.Score, &categoryID, &tradeMode, &room.AllowWalkthrough, &room.AllowPets, &room.AllowPetsEat, &room.HideWalls, &room.WallThickness, &room.FloorThickness, &room.ChatMode, &room.ChatWeight, &room.ChatSpeed, &room.ChatDistance, &room.ChatProtection, &moderationMute, &moderationKick, &moderationBan, &room.StaffPicked, &room.PublicRoom, &room.CreatedAt, &room.UpdatedAt, &deletedAt, &room.Version.Version)
	if err != nil {
		return roommodel.Room{}, err
	}
	room.CategoryID = int64Pointer(categoryID)
	room.PasswordHash = stringPointer(passwordHash)
	room.DeletedAt = timePointer(deletedAt)
	room.DoorMode = roommodel.DoorMode(doorMode)
	room.TradeMode = roommodel.TradeMode(tradeMode)
	room.ModerationMute = roommodel.ModerationPolicy(moderationMute)
	room.ModerationKick = roommodel.ModerationPolicy(moderationKick)
	room.ModerationBan = roommodel.ModerationPolicy(moderationBan)

	return room, nil
}

// stringPointer converts PostgreSQL text to an optional string.
func stringPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}

// scanRooms scans room rows.
func scanRooms(rows pgx.Rows) ([]roommodel.Room, error) {
	var rooms []roommodel.Room
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	return rooms, rows.Err()
}

// scanCategories scans room category rows.
func scanCategories(rows pgx.Rows) ([]roommodel.Category, error) {
	var categories []roommodel.Category
	for rows.Next() {
		var category roommodel.Category
		var deletedAt pgtype.Timestamptz
		err := rows.Scan(&category.ID, &category.Caption, &category.CaptionKey, &category.Visible, &category.Automatic, &category.AutomaticKey, &category.GlobalKey, &category.StaffOnly, &category.Order, &category.CreatedAt, &category.UpdatedAt, &deletedAt, &category.Version.Version)
		if err != nil {
			return nil, fmt.Errorf("scan room category: %w", err)
		}
		category.DeletedAt = timePointer(deletedAt)
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

// scanTags scans room tag rows.
func scanTags(rows pgx.Rows) ([]roommodel.Tag, error) {
	var tags []roommodel.Tag
	for rows.Next() {
		var tag roommodel.Tag
		if err := rows.Scan(&tag.RoomID, &tag.Value); err != nil {
			return nil, fmt.Errorf("scan room tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
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
