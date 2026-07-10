// Package repository contains PostgreSQL access for permission records.
package repository

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
)

// Reader reads permission groups, memberships, and grants.
type Reader interface {
	// ListGroups lists every active permission group.
	ListGroups(ctx context.Context) ([]permissionmodel.Group, error)
	// ListGroupsByPlayer lists one player's active groups by descending weight.
	ListGroupsByPlayer(ctx context.Context, playerID int64) ([]permissionmodel.Group, error)
	// FindGroupByID finds one active permission group.
	FindGroupByID(ctx context.Context, groupID int64) (permissionmodel.Group, bool, error)
	// FindGroupByName finds one active permission group by name.
	FindGroupByName(ctx context.Context, name string) (permissionmodel.Group, bool, error)
	// ListGroupNodes lists one group's grants.
	ListGroupNodes(ctx context.Context, groupID int64) ([]permissionmodel.Grant, error)
	// ListPlayerNodes lists one player's direct grants.
	ListPlayerNodes(ctx context.Context, playerID int64) ([]permissionmodel.Grant, error)
	// ListAffectedPlayerIDs lists players inheriting from a changed group.
	ListAffectedPlayerIDs(ctx context.Context, groupID int64) ([]int64, error)
}

// Writer mutates permission groups, memberships, and grants.
type Writer interface {
	// CreateGroup creates one permission group.
	CreateGroup(ctx context.Context, group permissionmodel.Group) (permissionmodel.Group, error)
	// UpdateGroup updates one permission group using optimistic locking.
	UpdateGroup(ctx context.Context, group permissionmodel.Group) (permissionmodel.Group, bool, error)
	// UpsertGroupNode creates or replaces one group grant.
	UpsertGroupNode(ctx context.Context, groupID int64, node permission.Node, allowed bool) error
	// DeleteGroupNode deletes one group grant.
	DeleteGroupNode(ctx context.Context, groupID int64, node permission.Node) error
	// AddPlayerToGroup adds one idempotent membership.
	AddPlayerToGroup(ctx context.Context, playerID int64, groupID int64) error
	// RemovePlayerFromGroup removes one membership.
	RemovePlayerFromGroup(ctx context.Context, playerID int64, groupID int64) error
	// UpsertPlayerNode creates or replaces one direct player grant.
	UpsertPlayerNode(ctx context.Context, playerID int64, node permission.Node, allowed bool) error
	// DeletePlayerNode deletes one direct player grant.
	DeletePlayerNode(ctx context.Context, playerID int64, node permission.Node) error
}

// Store reads and mutates permission persistence.
type Store interface {
	Reader
	Writer
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
