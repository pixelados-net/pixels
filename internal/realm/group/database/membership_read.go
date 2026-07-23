package database

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// playerGroupsSQL returns active memberships with a total favorite projection.
const playerGroupsSQL = `select ` + groupColumns + `,m.role,coalesce(preference.favorite_group_id=g.id,false)` + groupJoins + ` join social_group_members m on m.group_id=g.id left join player_social_group_preferences preference on preference.player_id=m.player_id where m.player_id=$1 and g.deactivated_at is null order by g.id`

// RoomFurnitureLinks returns linked furniture for one room generation.
func (repository *Repository) RoomFurnitureLinks(ctx context.Context, roomID int64) ([]grouprecord.GroupFurnitureLink, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select link.item_id,item.room_id,link.group_id from furniture_social_group_links link join furniture_items item on item.id=link.item_id join social_groups groups on groups.id=link.group_id and groups.deactivated_at is null where item.room_id=$1 and item.deleted_at is null order by link.item_id`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	links := make([]grouprecord.GroupFurnitureLink, 0)
	for rows.Next() {
		var link grouprecord.GroupFurnitureLink
		if err = rows.Scan(&link.ItemID, &link.RoomID, &link.GroupID); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

// PlayerInventoryFurnitureLinks returns active group links held in one player's inventory.
func (repository *Repository) PlayerInventoryFurnitureLinks(ctx context.Context, playerID int64) ([]grouprecord.GroupFurnitureLink, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select link.item_id,0,link.group_id from furniture_social_group_links link join furniture_items item on item.id=link.item_id join social_groups groups on groups.id=link.group_id and groups.deactivated_at is null where item.owner_player_id=$1 and item.room_id is null and item.deleted_at is null order by link.item_id`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	links := make([]grouprecord.GroupFurnitureLink, 0)
	for rows.Next() {
		var link grouprecord.GroupFurnitureLink
		if err = rows.Scan(&link.ItemID, &link.RoomID, &link.GroupID); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, rows.Err()
}

// Membership returns one active role.
func (repository *Repository) Membership(ctx context.Context, groupID int64, playerID int64) (grouprecord.Membership, bool, error) {
	var member grouprecord.Membership
	err := repository.executor(ctx).QueryRow(ctx, `select m.group_id,m.player_id,p.username,coalesce(profile.look,''),m.role,m.joined_at,m.updated_at,m.version from social_group_members m join social_groups g on g.id=m.group_id and g.deactivated_at is null join players p on p.id=m.player_id and p.deleted_at is null left join player_profiles profile on profile.player_id=p.id where m.group_id=$1 and m.player_id=$2`, groupID, playerID).Scan(&member.GroupID, &member.PlayerID, &member.Username, &member.Figure, &member.Role, &member.JoinedAt, &member.UpdatedAt, &member.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return grouprecord.Membership{}, false, nil
	}
	return member, err == nil, err
}

// Pending reports whether one membership request exists.
func (repository *Repository) Pending(ctx context.Context, groupID int64, playerID int64) (bool, error) {
	var pending bool
	err := repository.executor(ctx).QueryRow(ctx, `select exists(select 1 from social_group_requests request join social_groups groups on groups.id=request.group_id and groups.deactivated_at is null where request.group_id=$1 and request.player_id=$2)`, groupID, playerID).Scan(&pending)
	return pending, err
}

// PlayerGroups lists active memberships with favorite state.
func (repository *Repository) PlayerGroups(ctx context.Context, playerID int64) ([]grouprecord.PlayerGroup, error) {
	rows, err := repository.executor(ctx).Query(ctx, playerGroupsSQL, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]grouprecord.PlayerGroup, 0)
	for rows.Next() {
		group, scanErr := scanGroupWithTail(rows, nil)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, group)
	}
	return result, rows.Err()
}

// scanGroupWithTail scans shared columns plus role and favorite.
func scanGroupWithTail(row groupScanner, ignored any) (grouprecord.PlayerGroup, error) {
	var item grouprecord.PlayerGroup
	err := row.Scan(&item.Group.ID, &item.Group.OwnerPlayerID, &item.Group.OwnerName, &item.Group.Name, &item.Group.Description,
		&item.Group.HomeRoomID, &item.Group.HomeRoomName, &item.Group.State, &item.Group.CanMembersDecorate, &item.Group.ColorA, &item.Group.ColorB,
		&item.Group.ColorAHex, &item.Group.ColorBHex, &item.Group.BadgeCode, &item.Group.ForumEnabled, &item.Group.ReadPolicy,
		&item.Group.PostMessagePolicy, &item.Group.PostThreadPolicy, &item.Group.ModeratePolicy, &item.Group.MemberCount, &item.Group.PendingCount,
		&item.Group.ThreadCount, &item.Group.PostCount, &item.Group.CreatedAt, &item.Group.UpdatedAt, &item.Group.DeactivatedAt, &item.Group.Version,
		&item.Role, &item.Favorite)
	return item, err
}

// MemberPage returns a bounded filtered roster page.
func (repository *Repository) MemberPage(ctx context.Context, groupID int64, page int32, pageSize int32, query string, level int32) (grouprecord.MemberPage, error) {
	group, err := repository.requireGroup(ctx, groupID, false)
	if err != nil {
		return grouprecord.MemberPage{}, err
	}
	query = strings.TrimSpace(query)
	offset := page * pageSize
	result := grouprecord.MemberPage{Group: group, Page: page, PageSize: pageSize, Level: level, Query: query}
	if level == 2 {
		rows, total, requestErr := repository.requestPage(ctx, groupID, query, offset, pageSize)
		result.Members, result.Total = rows, total
		return result, requestErr
	}
	roleFilter := ``
	if level == 1 {
		roleFilter = ` and m.role in (0,1)`
	}
	if err = repository.executor(ctx).QueryRow(ctx, `select count(*) from social_group_members m join players p on p.id=m.player_id where m.group_id=$1 and ($2='' or lower(p.username) like '%'||lower($2)||'%')`+roleFilter, groupID, query).Scan(&result.Total); err != nil {
		return grouprecord.MemberPage{}, err
	}
	rows, err := repository.executor(ctx).Query(ctx, `select m.group_id,m.player_id,p.username,coalesce(profile.look,''),m.role,m.joined_at,m.updated_at,m.version from social_group_members m join players p on p.id=m.player_id left join player_profiles profile on profile.player_id=p.id where m.group_id=$1 and ($2='' or lower(p.username) like '%'||lower($2)||'%')`+roleFilter+` order by m.role,p.username,m.player_id offset $3 limit $4`, groupID, query, offset, pageSize)
	if err != nil {
		return grouprecord.MemberPage{}, err
	}
	defer rows.Close()
	for rows.Next() {
		member, scanErr := scanMembership(rows)
		if scanErr != nil {
			return grouprecord.MemberPage{}, scanErr
		}
		result.Members = append(result.Members, member)
	}
	return result, rows.Err()
}

// requestPage returns pending requests projected as wire members.
func (repository *Repository) requestPage(ctx context.Context, groupID int64, query string, offset int32, limit int32) ([]grouprecord.Membership, int32, error) {
	var total int32
	if err := repository.executor(ctx).QueryRow(ctx, `select count(*) from social_group_requests r join players p on p.id=r.player_id where r.group_id=$1 and ($2='' or lower(p.username) like '%'||lower($2)||'%')`, groupID, query).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := repository.executor(ctx).Query(ctx, `select r.group_id,r.player_id,p.username,coalesce(profile.look,''),3,r.requested_at,r.requested_at,1 from social_group_requests r join players p on p.id=r.player_id left join player_profiles profile on profile.player_id=p.id where r.group_id=$1 and ($2='' or lower(p.username) like '%'||lower($2)||'%') order by r.requested_at,r.player_id offset $3 limit $4`, groupID, query, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	members := make([]grouprecord.Membership, 0, limit)
	for rows.Next() {
		member, scanErr := scanMembership(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		members = append(members, member)
	}
	return members, total, rows.Err()
}

// scanMembership scans one roster row.
func scanMembership(row groupScanner) (grouprecord.Membership, error) {
	var member grouprecord.Membership
	err := row.Scan(&member.GroupID, &member.PlayerID, &member.Username, &member.Figure, &member.Role, &member.JoinedAt, &member.UpdatedAt, &member.Version)
	return member, err
}
