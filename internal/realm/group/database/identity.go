package database

import (
	"context"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// ReplaceBadge atomically replaces normalized parts and compiled code.
func (repository *Repository) ReplaceBadge(ctx context.Context, groupID int64, version int64, code string, parts []grouprecord.BadgePart) (grouprecord.Group, error) {
	var result grouprecord.Group
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		command, err := repository.executor(txCtx).Exec(txCtx, `update social_groups set badge_code=$3,updated_at=now(),version=version+1 where id=$1 and version=$2 and deactivated_at is null`, groupID, version, code)
		if err != nil {
			return mapConflict(err)
		}
		if command.RowsAffected() != 1 {
			return grouprecord.ErrConflict
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `delete from social_group_badge_parts where group_id=$1`, groupID); err != nil {
			return err
		}
		if err = repository.insertBadgeParts(txCtx, groupID, parts); err != nil {
			return err
		}
		if err = repository.audit(txCtx, groupID, "group.badge.updated", 0, version+1); err != nil {
			return err
		}
		var readErr error
		result, readErr = repository.requireGroup(txCtx, groupID, false)
		return readErr
	})
	return result, err
}

// insertBadgeParts inserts one normalized badge layer collection.
func (repository *Repository) insertBadgeParts(ctx context.Context, groupID int64, parts []grouprecord.BadgePart) error {
	for _, part := range parts {
		_, err := repository.executor(ctx).Exec(ctx, `insert into social_group_badge_parts(group_id,ordinal,kind,element_id,color_family,color_id,position) values($1,$2,$3,$4,$3,$5,$6)`, groupID, part.Ordinal, part.Kind, part.ElementID, part.ColorID, part.Position)
		if err != nil {
			return mapConflict(err)
		}
	}
	return nil
}

// BadgeParts returns stored normalized group badge layers.
func (repository *Repository) BadgeParts(ctx context.Context, groupID int64) ([]grouprecord.BadgePart, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select ordinal,kind,element_id,color_id,position from social_group_badge_parts where group_id=$1 order by ordinal`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	parts := make([]grouprecord.BadgePart, 0, 5)
	for rows.Next() {
		var part grouprecord.BadgePart
		if err = rows.Scan(&part.Ordinal, &part.Kind, &part.ElementID, &part.ColorID, &part.Position); err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	return parts, rows.Err()
}

// BadgeRegistry returns every enabled editor element and color.
func (repository *Repository) BadgeRegistry(ctx context.Context) ([]grouprecord.BadgeElement, []grouprecord.BadgeColor, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select kind,id,value_a,value_b,order_num from social_group_badge_elements where enabled order by kind,order_num,id`)
	if err != nil {
		return nil, nil, err
	}
	elements := make([]grouprecord.BadgeElement, 0)
	for rows.Next() {
		var element grouprecord.BadgeElement
		if err = rows.Scan(&element.Kind, &element.ID, &element.ValueA, &element.ValueB, &element.Order); err != nil {
			rows.Close()
			return nil, nil, err
		}
		elements = append(elements, element)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, nil, err
	}
	rows.Close()
	rows, err = repository.executor(ctx).Query(ctx, `select family,id,upper(hex),order_num from social_group_badge_colors where enabled order by family,order_num,id`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	colors := make([]grouprecord.BadgeColor, 0)
	for rows.Next() {
		var color grouprecord.BadgeColor
		if err = rows.Scan(&color.Family, &color.ID, &color.Hex, &color.Order); err != nil {
			return nil, nil, err
		}
		colors = append(colors, color)
	}
	return elements, colors, rows.Err()
}

// DeactivateGroup soft-deactivates group state and dependent preferences.
func (repository *Repository) DeactivateGroup(ctx context.Context, groupID int64, version int64) (grouprecord.Group, error) {
	var result grouprecord.Group
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		command, err := repository.executor(txCtx).Exec(txCtx, `update social_groups set deactivated_at=now(),updated_at=now(),version=version+1 where id=$1 and version=$2 and deactivated_at is null`, groupID, version)
		if err != nil {
			return err
		}
		if command.RowsAffected() != 1 {
			return grouprecord.ErrConflict
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `delete from room_social_groups where group_id=$1`, groupID); err != nil {
			return err
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update player_social_group_preferences set favorite_group_id=null,updated_at=now(),version=version+1 where favorite_group_id=$1`, groupID); err != nil {
			return err
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update furniture_social_group_links set group_id=null,updated_at=now(),version=version+1 where group_id=$1`, groupID); err != nil {
			return err
		}
		if err = repository.audit(txCtx, groupID, "group.deactivated", 0, version+1); err != nil {
			return err
		}
		result, err = repository.requireGroup(txCtx, groupID, true)
		return err
	})
	return result, err
}

// RestoreGroup validates and restores a retained group.
func (repository *Repository) RestoreGroup(ctx context.Context, groupID int64, version int64, roomID int64) (grouprecord.Group, error) {
	var result grouprecord.Group
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		var ownerID int64
		if err := repository.executor(txCtx).QueryRow(txCtx, `select coalesce(owner_player_id,0) from social_groups where id=$1 and version=$2 and deactivated_at is not null for update`, groupID, version).Scan(&ownerID); err != nil {
			return grouprecord.ErrConflict
		}
		if err := repository.LockEligibleRoom(txCtx, roomID, ownerID); err != nil {
			return err
		}
		if _, err := repository.executor(txCtx).Exec(txCtx, `insert into room_social_groups(room_id,group_id) values($1,$2)`, roomID, groupID); err != nil {
			return mapConflict(err)
		}
		if _, err := repository.executor(txCtx).Exec(txCtx, `update social_groups set home_room_id=$3,deactivated_at=null,updated_at=now(),version=version+1 where id=$1 and version=$2`, groupID, version, roomID); err != nil {
			return err
		}
		if err := repository.audit(txCtx, groupID, "group.restored", ownerID, version+1); err != nil {
			return err
		}
		var readErr error
		result, readErr = repository.requireGroup(txCtx, groupID, false)
		return readErr
	})
	return result, err
}

// TransferOwner atomically exchanges owner and target roles.
func (repository *Repository) TransferOwner(ctx context.Context, groupID int64, targetID int64, version int64) (grouprecord.Group, error) {
	var result grouprecord.Group
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		var ownerID int64
		if err := repository.executor(txCtx).QueryRow(txCtx, `select owner_player_id from social_groups where id=$1 and version=$2 and deactivated_at is null for update`, groupID, version).Scan(&ownerID); err != nil {
			return grouprecord.ErrConflict
		}
		command, err := repository.executor(txCtx).Exec(txCtx, `update social_group_members set role=case when player_id=$2 then 1 when player_id=$3 then 0 else role end,updated_at=now(),version=version+1 where group_id=$1 and player_id in ($2,$3)`, groupID, ownerID, targetID)
		if err != nil {
			return err
		}
		if command.RowsAffected() != 2 {
			return grouprecord.ErrConflict
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update social_groups set owner_player_id=$2,updated_at=now(),version=version+1 where id=$1`, groupID, targetID); err != nil {
			return err
		}
		if err = repository.audit(txCtx, groupID, "group.owner.transferred", targetID, version+1); err != nil {
			return err
		}
		result, err = repository.requireGroup(txCtx, groupID, false)
		return err
	})
	return result, err
}

// RebindRoom atomically replaces the home-room binding.
func (repository *Repository) RebindRoom(ctx context.Context, groupID int64, roomID int64, version int64) (grouprecord.Group, error) {
	var result grouprecord.Group
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		group, err := repository.requireGroup(txCtx, groupID, false)
		if err != nil || group.Version != version {
			return grouprecord.ErrConflict
		}
		if err = repository.LockEligibleRoom(txCtx, roomID, group.OwnerPlayerID); err != nil {
			return err
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `delete from room_social_groups where group_id=$1`, groupID); err != nil {
			return err
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `insert into room_social_groups(room_id,group_id) values($1,$2)`, roomID, groupID); err != nil {
			return mapConflict(err)
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update social_groups set home_room_id=$3,updated_at=now(),version=version+1 where id=$1 and version=$2`, groupID, version, roomID); err != nil {
			return err
		}
		if err = repository.audit(txCtx, groupID, "group.home_room.rebound", 0, version+1); err != nil {
			return err
		}
		result, err = repository.requireGroup(txCtx, groupID, false)
		return err
	})
	return result, err
}
