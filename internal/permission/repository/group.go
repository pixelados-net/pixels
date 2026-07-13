package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// groupColumns stores the shared permission group projection.
	groupColumns = `id, name, weight, prefix, prefix_color, parent_group_id, created_at, updated_at, deleted_at, version`
	// qualifiedGroupColumns stores the permission group projection for joined queries.
	qualifiedGroupColumns = `g.id, g.name, g.weight, g.prefix, g.prefix_color, g.parent_group_id, g.created_at, g.updated_at, g.deleted_at, g.version`
	// listGroupsSQL lists active permission groups by priority.
	listGroupsSQL = `select ` + groupColumns + ` from permission_groups where deleted_at is null order by weight desc, id`
	// listPlayerGroupsSQL lists one player's active permission groups.
	listPlayerGroupsSQL = `select ` + qualifiedGroupColumns + ` from permission_groups g join permission_player_groups pg on pg.group_id=g.id where pg.player_id=$1 and g.deleted_at is null order by g.weight desc, g.id`
	// findGroupByIDSQL finds one active permission group by id.
	findGroupByIDSQL = `select ` + groupColumns + ` from permission_groups where id=$1 and deleted_at is null`
	// findGroupByNameSQL finds one active permission group by name.
	findGroupByNameSQL = `select ` + groupColumns + ` from permission_groups where name=$1 and deleted_at is null`
	// createGroupSQL creates one permission group.
	createGroupSQL = `insert into permission_groups (name, weight, prefix, prefix_color, parent_group_id) values ($1,$2,$3,$4,$5) returning ` + groupColumns
	// updateGroupSQL updates one permission group with optimistic locking.
	updateGroupSQL = `update permission_groups set name=$2, weight=$3, prefix=$4, prefix_color=$5, parent_group_id=$6, updated_at=now(), version=version+1 where id=$1 and version=$7 and deleted_at is null returning ` + groupColumns
)

// ListGroups lists every active permission group.
func (repository *Repository) ListGroups(ctx context.Context) ([]permissionmodel.Group, error) {
	return repository.queryGroups(ctx, listGroupsSQL)
}

// ListGroupsByPlayer lists one player's active groups by descending weight.
func (repository *Repository) ListGroupsByPlayer(ctx context.Context, playerID int64) ([]permissionmodel.Group, error) {
	return repository.queryGroups(ctx, listPlayerGroupsSQL, playerID)
}

// FindGroupByID finds one active permission group.
func (repository *Repository) FindGroupByID(ctx context.Context, groupID int64) (permissionmodel.Group, bool, error) {
	return repository.queryGroup(ctx, findGroupByIDSQL, groupID)
}

// FindGroupByName finds one active permission group by name.
func (repository *Repository) FindGroupByName(ctx context.Context, name string) (permissionmodel.Group, bool, error) {
	return repository.queryGroup(ctx, findGroupByNameSQL, name)
}

// CreateGroup creates one permission group.
func (repository *Repository) CreateGroup(ctx context.Context, group permissionmodel.Group) (permissionmodel.Group, error) {
	created, err := scanGroup(postgres.ExecutorFor(ctx, repository.executor).QueryRow(ctx, createGroupSQL, group.Name, group.Weight, group.Prefix, group.PrefixColor, group.ParentGroupID))
	if err != nil {
		return permissionmodel.Group{}, fmt.Errorf("create permission group %q: %w", group.Name, err)
	}

	return created, nil
}

// UpdateGroup updates one permission group using optimistic locking.
func (repository *Repository) UpdateGroup(ctx context.Context, group permissionmodel.Group) (permissionmodel.Group, bool, error) {
	updated, err := scanGroup(postgres.ExecutorFor(ctx, repository.executor).QueryRow(ctx, updateGroupSQL,
		group.ID, group.Name, group.Weight, group.Prefix, group.PrefixColor, group.ParentGroupID, group.Version.Version))
	if errors.Is(err, pgx.ErrNoRows) {
		return permissionmodel.Group{}, false, nil
	}
	if err != nil {
		return permissionmodel.Group{}, false, fmt.Errorf("update permission group %d: %w", group.ID, err)
	}

	return updated, true, nil
}

// queryGroup scans one optional permission group.
func (repository *Repository) queryGroup(ctx context.Context, query string, arguments ...any) (permissionmodel.Group, bool, error) {
	group, err := scanGroup(postgres.ExecutorFor(ctx, repository.executor).QueryRow(ctx, query, arguments...))
	if errors.Is(err, pgx.ErrNoRows) {
		return permissionmodel.Group{}, false, nil
	}
	if err != nil {
		return permissionmodel.Group{}, false, err
	}

	return group, true, nil
}

// queryGroups scans a permission group collection.
func (repository *Repository) queryGroups(ctx context.Context, query string, arguments ...any) ([]permissionmodel.Group, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.executor).Query(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("query permission groups: %w", err)
	}
	defer rows.Close()

	return scanGroups(rows)
}
