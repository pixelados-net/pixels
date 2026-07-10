// Package service contains permission resolution and mutation behavior.
package service

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
)

// Checker resolves player permission capabilities.
type Checker interface {
	// HasPermission reports whether a player currently holds one concrete node.
	HasPermission(ctx context.Context, playerID int64, node permission.Node) (bool, error)
}

// DefaultAssigner assigns the configured default permission group.
type DefaultAssigner interface {
	// AssignDefaultGroup adds a player to the member group when it exists.
	AssignDefaultGroup(ctx context.Context, playerID int64) error
}

// Manager reads and mutates permission groups, memberships, and grants.
type Manager interface {
	Checker
	DefaultAssigner
	// Groups lists all active permission groups.
	Groups(ctx context.Context) ([]permissionmodel.Group, error)
	// EffectiveNodes returns registered nodes decided by player or group grants.
	EffectiveNodes(ctx context.Context, playerID int64) ([]ResolvedNode, error)
	// EffectivePerks returns allowed Nitro perk names.
	EffectivePerks(ctx context.Context, playerID int64) ([]string, error)
	// PrimaryGroup returns the player's highest-weight permission group.
	PrimaryGroup(ctx context.Context, playerID int64) (permissionmodel.Group, bool, error)
	// AffectedPlayerIDs lists players inheriting from one group.
	AffectedPlayerIDs(ctx context.Context, groupID int64) ([]int64, error)
	// CreateGroup creates one permission group.
	CreateGroup(ctx context.Context, params CreateGroupParams) (permissionmodel.Group, error)
	// UpdateGroup applies a partial permission group mutation.
	UpdateGroup(ctx context.Context, groupID int64, params UpdateGroupParams) (permissionmodel.Group, error)
	// GrantGroupNode creates or replaces one group grant.
	GrantGroupNode(ctx context.Context, groupID int64, node permission.Node, allowed bool) error
	// RevokeGroupNode removes one group grant.
	RevokeGroupNode(ctx context.Context, groupID int64, node permission.Node) error
	// AddPlayerToGroup creates one player membership.
	AddPlayerToGroup(ctx context.Context, playerID int64, groupID int64) error
	// RemovePlayerFromGroup removes one player membership.
	RemovePlayerFromGroup(ctx context.Context, playerID int64, groupID int64) error
	// GrantPlayerNode creates or replaces one direct player grant.
	GrantPlayerNode(ctx context.Context, playerID int64, node permission.Node, allowed bool) error
	// RevokePlayerNode removes one direct player grant.
	RevokePlayerNode(ctx context.Context, playerID int64, node permission.Node) error
}

// CreateGroupParams contains permission group creation fields.
type CreateGroupParams struct {
	// Name stores the unique group name.
	Name string
	// Weight stores group priority.
	Weight int32
	// Prefix stores a future chat prefix.
	Prefix string
	// PrefixColor stores a future chat prefix color.
	PrefixColor string
	// ParentGroupID identifies the optional inherited group.
	ParentGroupID *int64
}

// UpdateGroupParams contains optional permission group fields.
type UpdateGroupParams struct {
	// Name replaces the group name when present.
	Name *string
	// Weight replaces group priority when present.
	Weight *int32
	// Prefix replaces the chat prefix when present.
	Prefix *string
	// PrefixColor replaces the prefix color when present.
	PrefixColor *string
	// ParentGroupID replaces or clears the parent when present.
	ParentGroupID **int64
}

// ResolvedNode describes one effective node decision and its origin.
type ResolvedNode struct {
	// Node identifies the queried registered node.
	Node permission.Node
	// Allowed stores the resolved decision.
	Allowed bool
	// Source identifies player override or deciding group.
	Source string
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
