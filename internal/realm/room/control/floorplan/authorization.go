package floorplan

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

var (
	// ErrAccessDenied reports an actor without floor plan capability.
	ErrAccessDenied = errors.New("room floor plan access denied")
)

// RightsChecker resolves persistent room build rights.
type RightsChecker interface {
	// HasRights reports whether a player holds room build rights.
	HasRights(context.Context, int64, int64) (bool, error)
}

// Nodes contains floor plan permission nodes.
type Nodes struct {
	// OwnEdit authorizes local owner or rights-holder editing.
	OwnEdit permission.Node
	// AnyEdit authorizes editing any room.
	AnyEdit permission.Node
}

// Authorizer resolves floor plan editing capability.
type Authorizer struct {
	// permissions resolves global capabilities.
	permissions permissionservice.Checker
	// rights resolves local build rights.
	rights RightsChecker
	// nodes stores floor plan permission nodes.
	nodes Nodes
}

// NewAuthorizer creates a floor plan authorizer.
func NewAuthorizer(permissions permissionservice.Checker, rights RightsChecker, nodes Nodes) *Authorizer {
	return &Authorizer{permissions: permissions, rights: rights, nodes: nodes}
}

// Authorize requires floor plan editing capability for one room.
func (authorizer *Authorizer) Authorize(ctx context.Context, room roommodel.Room, actorID int64) error {
	allowed, err := authorizer.has(ctx, actorID, authorizer.nodes.AnyEdit)
	if err != nil || allowed {
		return authorizationResult(allowed, err)
	}
	local := room.OwnerPlayerID == actorID
	if !local && authorizer.rights != nil {
		local, err = authorizer.rights.HasRights(ctx, room.ID, actorID)
		if err != nil {
			return err
		}
	}
	if !local {
		return ErrAccessDenied
	}
	allowed, err = authorizer.has(ctx, actorID, authorizer.nodes.OwnEdit)

	return authorizationResult(allowed, err)
}

// has resolves one permission node.
func (authorizer *Authorizer) has(ctx context.Context, actorID int64, node permission.Node) (bool, error) {
	if authorizer == nil || authorizer.permissions == nil || actorID <= 0 || node == "" {
		return false, nil
	}

	return authorizer.permissions.HasPermission(ctx, actorID, node)
}

// authorizationResult maps permission resolution to a domain error.
func authorizationResult(allowed bool, err error) error {
	if err != nil {
		return err
	}
	if !allowed {
		return ErrAccessDenied
	}

	return nil
}
