package repository

import (
	"context"
	"fmt"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

const (
	// deleteRoomTagsSQL removes room tags before replacement.
	deleteRoomTagsSQL = `delete from room_tags where room_id = $1`

	// insertRoomTagSQL inserts one room tag.
	insertRoomTagSQL = `insert into room_tags (room_id, tag) values ($1, $2) on conflict do nothing`

	// listRoomTagsSQL reads room tags.
	listRoomTagsSQL = `select room_id, tag from room_tags where room_id = $1 order by tag asc`
)

// ListRoomTags lists tags for a room.
func (repository *Repository) ListRoomTags(ctx context.Context, roomID int64) ([]roommodel.Tag, error) {
	rows, err := repository.executor.Query(ctx, listRoomTagsSQL, roomID)
	if err != nil {
		return nil, fmt.Errorf("list room tags: %w", err)
	}
	defer rows.Close()

	return scanTags(rows)
}

// ReplaceRoomTags replaces tags for a room.
func (repository *Repository) ReplaceRoomTags(ctx context.Context, roomID int64, tags []string) error {
	if _, err := repository.executor.Exec(ctx, deleteRoomTagsSQL, roomID); err != nil {
		return fmt.Errorf("delete room tags: %w", err)
	}

	for _, tag := range tags {
		if _, err := repository.executor.Exec(ctx, insertRoomTagSQL, roomID, tag); err != nil {
			return fmt.Errorf("insert room tag: %w", err)
		}
	}

	return nil
}
