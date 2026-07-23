package database

import (
	"context"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// returnHQFurnitureSQL atomically identifies and returns one bounded headquarters furniture batch.
const returnHQFurnitureSQL = `with targets as materialized (
select item.id as item_id,item.room_id,item.owner_player_id,(item.wall_position is not null) as wall
from furniture_items item join social_groups groups on groups.home_room_id=item.room_id
where groups.id=$1 and item.owner_player_id=$2 and item.deleted_at is null
order by item.id limit $3 for update of item
), updated as (
update furniture_items item set room_id=null,x=null,y=null,z=null,wall_position=null,updated_at=now(),version=item.version+1
from targets where item.id=targets.item_id
returning targets.item_id,targets.room_id,targets.owner_player_id,targets.wall
)
select item_id,room_id,owner_player_id,wall from updated order by item_id`

func (repository *Repository) ChangeRole(ctx context.Context, groupID int64, playerID int64, role grouprecord.Role) (grouprecord.Membership, error) {
	if role != grouprecord.Admin && role != grouprecord.Member {
		return grouprecord.Membership{}, grouprecord.ErrInvalid
	}
	var member grouprecord.Membership
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		command, err := repository.executor(txCtx).Exec(txCtx, `update social_group_members set role=$3,updated_at=now(),version=version+1 where group_id=$1 and player_id=$2 and role<>0`, groupID, playerID, role)
		if err != nil {
			return err
		}
		if command.RowsAffected() != 1 {
			return grouprecord.ErrForbidden
		}
		member, _, err = repository.Membership(txCtx, groupID, playerID)
		if err != nil {
			return err
		}
		return repository.audit(txCtx, groupID, "group.member.role_changed", playerID, member.Version)
	})
	return member, err
}

// FurnitureCount counts target-owned furniture in the group headquarters.
func (repository *Repository) FurnitureCount(ctx context.Context, groupID int64, playerID int64) (int, error) {
	var count int
	err := repository.executor(ctx).QueryRow(ctx, `select count(*) from furniture_items item join social_groups groups on groups.home_room_id=item.room_id where groups.id=$1 and item.owner_player_id=$2 and item.deleted_at is null`, groupID, playerID).Scan(&count)
	return count, err
}

// RemoveMember removes membership, favorite, and returns HQ furniture atomically.
func (repository *Repository) RemoveMember(ctx context.Context, groupID int64, playerID int64, cleanupLimit int) (grouprecord.FurnitureReturn, error) {
	var returned grouprecord.FurnitureReturn
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		if cleanupLimit < 0 {
			return grouprecord.ErrInvalid
		}
		member, found, err := repository.Membership(txCtx, groupID, playerID)
		if err != nil {
			return err
		}
		if !found {
			_, err = repository.DeclineRequest(txCtx, groupID, playerID)
			return err
		}
		if member.Role == grouprecord.Owner {
			return grouprecord.ErrForbidden
		}
		rows, err := repository.executor(txCtx).Query(txCtx, returnHQFurnitureSQL, groupID, playerID, cleanupLimit+1)
		if err != nil {
			return err
		}
		for rows.Next() {
			if len(returned.Items) == 0 {
				capacity := cleanupLimit
				if capacity < 1 {
					capacity = 1
				}
				if capacity > 16 {
					capacity = 16
				}
				returned.Items = make([]grouprecord.ReturnedFurniture, 0, capacity)
			}
			var item grouprecord.ReturnedFurniture
			if err = rows.Scan(&item.ItemID, &returned.RoomID, &item.OwnerPlayerID, &item.Wall); err != nil {
				rows.Close()
				return err
			}
			returned.Items = append(returned.Items, item)
		}
		if err = rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		if returned.Count() > cleanupLimit {
			return grouprecord.ErrLimit
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `delete from social_group_members where group_id=$1 and player_id=$2 and role<>0`, groupID, playerID); err != nil {
			return err
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update player_social_group_preferences set favorite_group_id=null,updated_at=now(),version=version+1 where player_id=$2 and favorite_group_id=$1`, groupID, playerID); err != nil {
			return err
		}
		if _, err = repository.executor(txCtx).Exec(txCtx, `update social_groups set member_count=greatest(member_count-1,0),updated_at=now(),version=version+1 where id=$1`, groupID); err != nil {
			return err
		}
		return repository.audit(txCtx, groupID, "group.member.removed", playerID, 0)
	})
	return returned, err
}

// SetFavorite validates membership and replaces one player preference.
func (repository *Repository) SetFavorite(ctx context.Context, playerID int64, groupID *int64) error {
	return repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		if groupID != nil {
			if _, found, err := repository.Membership(txCtx, *groupID, playerID); err != nil {
				return err
			} else if !found {
				return grouprecord.ErrForbidden
			}
		}
		var previousID int64
		_ = repository.executor(txCtx).QueryRow(txCtx, `select coalesce(favorite_group_id,0) from player_social_group_preferences where player_id=$1 for update`, playerID).Scan(&previousID)
		if _, err := repository.executor(txCtx).Exec(txCtx, `insert into player_social_group_preferences(player_id,favorite_group_id) values($1,$2) on conflict(player_id) do update set favorite_group_id=excluded.favorite_group_id,updated_at=now(),version=player_social_group_preferences.version+1`, playerID, groupID); err != nil {
			return err
		}
		auditGroupID := previousID
		if groupID != nil {
			auditGroupID = *groupID
		}
		if auditGroupID <= 0 {
			return nil
		}
		return repository.audit(txCtx, auditGroupID, "group.favorite.changed", playerID, 0)
	})
}

// LinkFurniture links granted inventory items to one active group.
func (repository *Repository) LinkFurniture(ctx context.Context, groupID int64, itemIDs []int64) error {
	if len(itemIDs) == 0 {
		return nil
	}
	command, err := repository.executor(ctx).Exec(ctx, `insert into furniture_social_group_links(item_id,group_id) select item.id,$1 from furniture_items item where item.id=any($2) and item.deleted_at is null on conflict(item_id) do update set group_id=excluded.group_id,updated_at=now(),version=furniture_social_group_links.version+1`, groupID, itemIDs)
	if err != nil {
		return err
	}
	if command.RowsAffected() != int64(len(itemIDs)) {
		return grouprecord.ErrConflict
	}
	return nil
}

// EnableForum activates one active group's forum entitlement idempotently.
func (repository *Repository) EnableForum(ctx context.Context, groupID int64) (bool, error) {
	command, err := repository.executor(ctx).Exec(ctx, `update social_groups set forum_enabled=true,updated_at=now(),version=version+1 where id=$1 and deactivated_at is null and not forum_enabled`, groupID)
	return command.RowsAffected() == 1, err
}

// Requests lists bounded pending requests.
func (repository *Repository) Requests(ctx context.Context, groupID int64, offset int, limit int) ([]grouprecord.Request, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select r.group_id,r.player_id,p.username,coalesce(profile.look,''),r.requested_at from social_group_requests r join players p on p.id=r.player_id and p.deleted_at is null left join player_profiles profile on profile.player_id=p.id where r.group_id=$1 order by r.requested_at,r.player_id offset $2 limit $3`, groupID, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requests := make([]grouprecord.Request, 0, limit)
	for rows.Next() {
		var request grouprecord.Request
		if err = rows.Scan(&request.GroupID, &request.PlayerID, &request.Username, &request.Figure, &request.RequestedAt); err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	return requests, rows.Err()
}
