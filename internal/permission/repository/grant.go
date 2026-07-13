package repository

import (
	"context"
	"fmt"

	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// listGroupNodesSQL lists one group's grants.
	listGroupNodesSQL = `select node, allowed from permission_group_nodes where group_id=$1 order by node`
	// listPlayerNodesSQL lists one player's direct grants.
	listPlayerNodesSQL = `select node, allowed from permission_player_nodes where player_id=$1 order by node`
	// upsertGroupNodeSQL creates or replaces one group grant.
	upsertGroupNodeSQL = `insert into permission_group_nodes (group_id,node,allowed) values ($1,$2,$3) on conflict (group_id,node) do update set allowed=excluded.allowed`
	// deleteGroupNodeSQL deletes one group grant.
	deleteGroupNodeSQL = `delete from permission_group_nodes where group_id=$1 and node=$2`
	// addPlayerGroupSQL creates one idempotent group membership.
	addPlayerGroupSQL = `insert into permission_player_groups (player_id,group_id) values ($1,$2) on conflict do nothing`
	// removePlayerGroupSQL removes one group membership.
	removePlayerGroupSQL = `delete from permission_player_groups where player_id=$1 and group_id=$2`
	// upsertPlayerNodeSQL creates or replaces one direct player grant.
	upsertPlayerNodeSQL = `insert into permission_player_nodes (player_id,node,allowed) values ($1,$2,$3) on conflict (player_id,node) do update set allowed=excluded.allowed`
	// deletePlayerNodeSQL deletes one direct player grant.
	deletePlayerNodeSQL = `delete from permission_player_nodes where player_id=$1 and node=$2`
	// affectedPlayersSQL lists direct and descendant group members.
	affectedPlayersSQL = `with recursive descendants as (select id from permission_groups where id=$1 and deleted_at is null union all select g.id from permission_groups g join descendants d on g.parent_group_id=d.id where g.deleted_at is null) select distinct pg.player_id from permission_player_groups pg join descendants d on d.id=pg.group_id order by pg.player_id`
)

// ListGroupNodes lists one group's grants.
func (repository *Repository) ListGroupNodes(ctx context.Context, groupID int64) ([]permissionmodel.Grant, error) {
	return repository.queryGrants(ctx, listGroupNodesSQL, groupID)
}

// ListPlayerNodes lists one player's direct grants.
func (repository *Repository) ListPlayerNodes(ctx context.Context, playerID int64) ([]permissionmodel.Grant, error) {
	return repository.queryGrants(ctx, listPlayerNodesSQL, playerID)
}

// ListAffectedPlayerIDs lists players inheriting from a changed group.
func (repository *Repository) ListAffectedPlayerIDs(ctx context.Context, groupID int64) ([]int64, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.executor).Query(ctx, affectedPlayersSQL, groupID)
	if err != nil {
		return nil, fmt.Errorf("list players affected by permission group %d: %w", groupID, err)
	}
	defer rows.Close()

	playerIDs := make([]int64, 0)
	for rows.Next() {
		var playerID int64
		if err := rows.Scan(&playerID); err != nil {
			return nil, err
		}
		playerIDs = append(playerIDs, playerID)
	}

	return playerIDs, rows.Err()
}

// UpsertGroupNode creates or replaces one group grant.
func (repository *Repository) UpsertGroupNode(ctx context.Context, groupID int64, node permission.Node, allowed bool) error {
	return repository.execute(ctx, upsertGroupNodeSQL, groupID, node, allowed)
}

// DeleteGroupNode deletes one group grant.
func (repository *Repository) DeleteGroupNode(ctx context.Context, groupID int64, node permission.Node) error {
	return repository.execute(ctx, deleteGroupNodeSQL, groupID, node)
}

// AddPlayerToGroup adds one idempotent membership.
func (repository *Repository) AddPlayerToGroup(ctx context.Context, playerID int64, groupID int64) error {
	return repository.execute(ctx, addPlayerGroupSQL, playerID, groupID)
}

// RemovePlayerFromGroup removes one membership.
func (repository *Repository) RemovePlayerFromGroup(ctx context.Context, playerID int64, groupID int64) error {
	return repository.execute(ctx, removePlayerGroupSQL, playerID, groupID)
}

// UpsertPlayerNode creates or replaces one direct player grant.
func (repository *Repository) UpsertPlayerNode(ctx context.Context, playerID int64, node permission.Node, allowed bool) error {
	return repository.execute(ctx, upsertPlayerNodeSQL, playerID, node, allowed)
}

// DeletePlayerNode deletes one direct player grant.
func (repository *Repository) DeletePlayerNode(ctx context.Context, playerID int64, node permission.Node) error {
	return repository.execute(ctx, deletePlayerNodeSQL, playerID, node)
}

// queryGrants scans a grant collection.
func (repository *Repository) queryGrants(ctx context.Context, query string, identifier int64) ([]permissionmodel.Grant, error) {
	rows, err := postgres.ExecutorFor(ctx, repository.executor).Query(ctx, query, identifier)
	if err != nil {
		return nil, fmt.Errorf("query permission grants: %w", err)
	}
	defer rows.Close()

	return scanGrants(rows)
}

// execute runs one permission mutation statement.
func (repository *Repository) execute(ctx context.Context, query string, arguments ...any) error {
	if _, err := postgres.ExecutorFor(ctx, repository.executor).Exec(ctx, query, arguments...); err != nil {
		return fmt.Errorf("mutate permission persistence: %w", err)
	}

	return nil
}
