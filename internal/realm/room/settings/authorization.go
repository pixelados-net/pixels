// Package settings authorizes room configuration management.
package settings

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
)

var (
	// ErrAccessDenied reports an actor without room settings capability.
	ErrAccessDenied = errors.New("room settings access denied")
	// ErrClubRequired reports a restricted room setting without active club access.
	ErrClubRequired = errors.New("active club subscription required")
)

// Nodes contains room settings permission nodes.
type Nodes struct {
	// OwnManage authorizes local owner or rights-holder management.
	OwnManage permission.Node
	// AnyManage authorizes global room settings management.
	AnyManage permission.Node
	// OwnPolicyManage authorizes owners to change room moderation policy.
	OwnPolicyManage permission.Node
	// AnyPolicyManage authorizes global room moderation policy changes.
	AnyPolicyManage permission.Node
}

// Authorizer resolves room settings capabilities.
type Authorizer struct {
	// permissions resolves global capabilities.
	permissions permissionservice.Checker
	// nodes stores settings permission nodes.
	nodes Nodes
}

// New creates a room settings authorizer.
func New(permissions permissionservice.Checker, nodes Nodes) *Authorizer {
	return &Authorizer{permissions: permissions, nodes: nodes}
}

// CanManage reports whether an actor may manage one room.
func (authorizer *Authorizer) CanManage(ctx context.Context, room roommodel.Room, actorID int64) (bool, error) {
	if actorID <= 0 || room.ID <= 0 {
		return false, ErrAccessDenied
	}
	allowed, err := authorizer.has(ctx, actorID, authorizer.nodes.AnyManage)
	if err != nil || allowed {
		return allowed, err
	}
	if room.OwnerPlayerID != actorID {
		return false, nil
	}

	return authorizer.has(ctx, actorID, authorizer.nodes.OwnManage)
}

// AuthorizePolicy requires room moderation-policy management capability.
func (authorizer *Authorizer) AuthorizePolicy(ctx context.Context, room roommodel.Room, actorID int64) error {
	allowed, err := authorizer.has(ctx, actorID, authorizer.nodes.AnyPolicyManage)
	if err != nil {
		return err
	}
	if !allowed && room.OwnerPlayerID == actorID {
		allowed, err = authorizer.has(ctx, actorID, authorizer.nodes.OwnPolicyManage)
	}
	if err != nil {
		return err
	}
	if !allowed {
		return ErrAccessDenied
	}

	return nil
}

// Authorize requires room settings capability.
func (authorizer *Authorizer) Authorize(ctx context.Context, room roommodel.Room, actorID int64) error {
	allowed, err := authorizer.CanManage(ctx, room, actorID)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrAccessDenied
	}

	return nil
}

// CanManageAny reports whether an actor has global settings capability.
func (authorizer *Authorizer) CanManageAny(ctx context.Context, actorID int64) (bool, error) {
	return authorizer.has(ctx, actorID, authorizer.nodes.AnyManage)
}

// has resolves one permission node.
func (authorizer *Authorizer) has(ctx context.Context, actorID int64, node permission.Node) (bool, error) {
	if authorizer == nil || authorizer.permissions == nil || node == "" {
		return false, nil
	}

	return authorizer.permissions.HasPermission(ctx, actorID, node)
}
