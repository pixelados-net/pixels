package database

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// groupColumns contains the shared resolved social-group read model.
const groupColumns = `g.id,coalesce(g.owner_player_id,0),coalesce(owner.username,''),g.name,g.description,coalesce(g.home_room_id,0),coalesce(room.name,''),g.state,g.can_members_decorate,g.color_a,g.color_b,coalesce(ca.hex,''),coalesce(cb.hex,''),g.badge_code,g.forum_enabled,g.forum_read_policy,g.forum_post_message_policy,g.forum_post_thread_policy,g.forum_moderate_policy,g.member_count,g.pending_count,g.thread_count,g.post_count,g.created_at,g.updated_at,g.deactivated_at,g.version`

// groupJoins contains shared display-data joins.
const groupJoins = ` from social_groups g left join players owner on owner.id=g.owner_player_id left join rooms room on room.id=g.home_room_id left join social_group_badge_colors ca on ca.family=2 and ca.id=g.color_a left join social_group_badge_colors cb on cb.family=2 and cb.id=g.color_b `

// RoomGroup returns one active room binding and immutable group metadata.
func (repository *Repository) RoomGroup(ctx context.Context, roomID int64) (grouprecord.Group, bool, error) {
	group, err := scanGroup(repository.executor(ctx).QueryRow(ctx, `select `+groupColumns+groupJoins+` join room_social_groups binding on binding.group_id=g.id where binding.room_id=$1 and g.deactivated_at is null`, roomID))
	if errors.Is(err, pgx.ErrNoRows) {
		return grouprecord.Group{}, false, nil
	}
	return group, err == nil, err
}

// RoomGroups returns active group bindings for one bounded room batch.
func (repository *Repository) RoomGroups(ctx context.Context, roomIDs []int64) ([]grouprecord.RoomBinding, error) {
	if len(roomIDs) == 0 {
		return nil, nil
	}
	rows, err := repository.executor(ctx).Query(ctx, `select binding.room_id,`+groupColumns+groupJoins+` join room_social_groups binding on binding.group_id=g.id where binding.room_id=any($1) and g.deactivated_at is null order by binding.room_id`, roomIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bindings := make([]grouprecord.RoomBinding, 0, len(roomIDs))
	for rows.Next() {
		var binding grouprecord.RoomBinding
		values := make([]any, 0, 28)
		values = append(values, &binding.RoomID)
		values = append(values, groupScanTargets(&binding.Group)...)
		if err = rows.Scan(values...); err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, rows.Err()
}

// groupScanTargets returns shared destinations for one group projection.
func groupScanTargets(group *grouprecord.Group) []any {
	return []any{&group.ID, &group.OwnerPlayerID, &group.OwnerName, &group.Name, &group.Description,
		&group.HomeRoomID, &group.HomeRoomName, &group.State, &group.CanMembersDecorate, &group.ColorA, &group.ColorB,
		&group.ColorAHex, &group.ColorBHex, &group.BadgeCode, &group.ForumEnabled, &group.ReadPolicy,
		&group.PostMessagePolicy, &group.PostThreadPolicy, &group.ModeratePolicy, &group.MemberCount, &group.PendingCount,
		&group.ThreadCount, &group.PostCount, &group.CreatedAt, &group.UpdatedAt, &group.DeactivatedAt, &group.Version}
}

// groupScanner is implemented by PostgreSQL row values.
type groupScanner interface{ Scan(...any) error }

// scanGroup maps the shared group read model.
func scanGroup(row groupScanner) (grouprecord.Group, error) {
	var group grouprecord.Group
	err := row.Scan(groupScanTargets(&group)...)
	return group, err
}

// EligibleRooms lists active unbound non-template rooms owned by a player.
func (repository *Repository) EligibleRooms(ctx context.Context, playerID int64) ([]grouprecord.EligibleRoom, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select r.id,r.name from rooms r left join room_social_groups binding on binding.room_id=r.id where r.owner_player_id=$1 and r.deleted_at is null and not r.is_bundle_template and binding.room_id is null order by lower(r.name),r.id`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]grouprecord.EligibleRoom, 0)
	for rows.Next() {
		var room grouprecord.EligibleRoom
		if err = rows.Scan(&room.ID, &room.Name); err != nil {
			return nil, err
		}
		result = append(result, room)
	}
	return result, rows.Err()
}

// LockEligibleRoom validates and locks an active unbound room owned by a player.
func (repository *Repository) LockEligibleRoom(ctx context.Context, roomID int64, ownerID int64) error {
	var lockedID int64
	err := repository.executor(ctx).QueryRow(ctx, `select r.id from rooms r left join room_social_groups binding on binding.room_id=r.id where r.id=$1 and r.owner_player_id=$2 and r.deleted_at is null and not r.is_bundle_template and binding.room_id is null for update of r`, roomID, ownerID).Scan(&lockedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return grouprecord.ErrConflict
	}
	return err
}

// CountOwned counts active groups owned by one player.
func (repository *Repository) CountOwned(ctx context.Context, playerID int64) (int, error) {
	var count int
	err := repository.executor(ctx).QueryRow(ctx, `select count(*) from social_groups where owner_player_id=$1 and deactivated_at is null`, playerID).Scan(&count)
	return count, err
}

// CountMemberships counts active memberships held by one player.
func (repository *Repository) CountMemberships(ctx context.Context, playerID int64) (int, error) {
	var count int
	err := repository.executor(ctx).QueryRow(ctx, `select count(*) from social_group_members members join social_groups groups on groups.id=members.group_id and groups.deactivated_at is null where members.player_id=$1`, playerID).Scan(&count)
	return count, err
}

// InsertGroup creates group, owner membership, badge parts, and room binding.
func (repository *Repository) InsertGroup(ctx context.Context, params grouprecord.CreateParams) (grouprecord.Group, error) {
	executor := repository.executor(ctx)
	var groupID int64
	err := executor.QueryRow(ctx, `insert into social_groups(owner_player_id,name,description,home_room_id,state,color_a,color_b,badge_code,member_count) values($1,$2,$3,$4,$5,$6,$7,$8,1) returning id`, params.OwnerPlayerID, params.Name, params.Description, params.HomeRoomID, params.State, params.ColorA, params.ColorB, params.BadgeCode).Scan(&groupID)
	if err != nil {
		return grouprecord.Group{}, mapConflict(err)
	}
	if _, err = executor.Exec(ctx, `insert into social_group_members(group_id,player_id,role) values($1,$2,0)`, groupID, params.OwnerPlayerID); err != nil {
		return grouprecord.Group{}, mapConflict(err)
	}
	if _, err = executor.Exec(ctx, `insert into room_social_groups(room_id,group_id) values($1,$2)`, params.HomeRoomID, groupID); err != nil {
		return grouprecord.Group{}, mapConflict(err)
	}
	if err = repository.insertBadgeParts(ctx, groupID, params.BadgeParts); err != nil {
		return grouprecord.Group{}, err
	}
	if err = repository.audit(ctx, groupID, "group.created", params.OwnerPlayerID, 1); err != nil {
		return grouprecord.Group{}, err
	}
	return repository.requireGroup(ctx, groupID, false)
}

// Group returns one group and resolved display metadata.
func (repository *Repository) Group(ctx context.Context, groupID int64, includeDeactivated bool) (grouprecord.Group, bool, error) {
	query := `select ` + groupColumns + groupJoins + ` where g.id=$1`
	if !includeDeactivated {
		query += ` and g.deactivated_at is null`
	}
	group, err := scanGroup(repository.executor(ctx).QueryRow(ctx, query, groupID))
	if errors.Is(err, pgx.ErrNoRows) {
		return grouprecord.Group{}, false, nil
	}
	return group, err == nil, err
}

// requireGroup returns one group or a domain not-found error.
func (repository *Repository) requireGroup(ctx context.Context, groupID int64, includeDeactivated bool) (grouprecord.Group, error) {
	group, found, err := repository.Group(ctx, groupID, includeDeactivated)
	if err != nil {
		return grouprecord.Group{}, err
	}
	if !found {
		return grouprecord.Group{}, grouprecord.ErrNotFound
	}
	return group, nil
}

// Groups lists groups for administration.
func (repository *Repository) Groups(ctx context.Context, filter grouprecord.GroupFilter) ([]grouprecord.Group, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select `+groupColumns+groupJoins+` where ($1='' or lower(g.name) like '%'||lower($1)||'%') and ($2=0 or g.owner_player_id=$2) and ($3=0 or g.home_room_id=$3) and ($4::smallint is null or g.state=$4) and ($5::boolean is null or g.forum_enabled=$5) and ($6::boolean is null or (g.deactivated_at is null)=$6) order by g.id desc offset $7 limit $8`, strings.TrimSpace(filter.Query), filter.OwnerPlayerID, filter.HomeRoomID, filter.State, filter.ForumEnabled, filter.Active, filter.Offset, filter.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := make([]grouprecord.Group, 0, filter.Limit)
	for rows.Next() {
		group, scanErr := scanGroup(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

// PopularGroups lists active groups with valid headquarters by descending member count.
func (repository *Repository) PopularGroups(ctx context.Context, limit int) ([]grouprecord.Group, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select `+groupColumns+groupJoins+` where g.deactivated_at is null and g.home_room_id is not null and room.deleted_at is null and not room.is_bundle_template order by g.member_count desc,g.id desc limit $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := make([]grouprecord.Group, 0, limit)
	for rows.Next() {
		group, scanErr := scanGroup(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

// UpdateGroup applies an optimistic metadata patch.
func (repository *Repository) UpdateGroup(ctx context.Context, groupID int64, version int64, patch grouprecord.GroupPatch) (grouprecord.Group, error) {
	var result grouprecord.Group
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		command, err := repository.executor(txCtx).Exec(txCtx, `update social_groups set name=coalesce($3,name),description=coalesce($4,description),state=coalesce($5,state),can_members_decorate=coalesce($6,can_members_decorate),color_a=coalesce($7,color_a),color_b=coalesce($8,color_b),forum_enabled=coalesce($9,forum_enabled),forum_read_policy=coalesce($10,forum_read_policy),forum_post_message_policy=coalesce($11,forum_post_message_policy),forum_post_thread_policy=coalesce($12,forum_post_thread_policy),forum_moderate_policy=coalesce($13,forum_moderate_policy),updated_at=now(),version=version+1 where id=$1 and version=$2 and deactivated_at is null`, groupID, version, patch.Name, patch.Description, patch.State, patch.CanMembersDecorate, patch.ColorA, patch.ColorB, patch.ForumEnabled, patch.ReadPolicy, patch.PostMessagePolicy, patch.PostThreadPolicy, patch.ModeratePolicy)
		if err != nil {
			return mapConflict(err)
		}
		if command.RowsAffected() != 1 {
			return grouprecord.ErrConflict
		}
		if err = repository.audit(txCtx, groupID, "group.updated", 0, version+1); err != nil {
			return err
		}
		result, err = repository.requireGroup(txCtx, groupID, false)
		return err
	})
	return result, err
}

// mapConflict maps PostgreSQL constraint failures to a stable domain conflict.
func mapConflict(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "violates check constraint") || strings.Contains(err.Error(), "violates foreign key") {
		return grouprecord.ErrConflict
	}
	return err
}
