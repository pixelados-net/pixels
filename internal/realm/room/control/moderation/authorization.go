package moderation

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// actionPolicy stores the policy and permission nodes for one action.
type actionPolicy struct {
	// policy stores the room-scoped moderation setting.
	policy roommodel.ModerationPolicy
	// ownNode stores the local actor capability.
	ownNode permission.Node
	// anyNode stores the global staff capability.
	anyNode permission.Node
}

// CanModerate reports whether an actor may perform a room action before selecting a target.
func (service *Service) CanModerate(ctx context.Context, room roommodel.Room, actorID int64, action moderationmodel.Action) (bool, error) {
	if actorID <= 0 || room.ID <= 0 {
		return false, ErrInvalidIdentity
	}
	policy := service.policy(room, action)
	allowed, err := service.hasPermission(ctx, actorID, policy.anyNode)
	if err != nil || allowed {
		return allowed, err
	}

	return service.localAllowed(ctx, room, actorID, policy)
}

// authorize applies staff, owner, rights, local-node, and target protection policy.
func (service *Service) authorize(ctx context.Context, roomID int64, actorID int64, targetID int64, action moderationmodel.Action) (roommodel.Room, error) {
	if roomID <= 0 || actorID <= 0 || targetID <= 0 {
		return roommodel.Room{}, ErrInvalidIdentity
	}
	if actorID == targetID {
		return roommodel.Room{}, ErrSelfTarget
	}
	room, found, err := service.rooms.FindByID(ctx, roomID)
	if err != nil {
		return roommodel.Room{}, err
	}
	if !found {
		return roommodel.Room{}, ErrRoomNotFound
	}
	if room.OwnerPlayerID == targetID {
		return roommodel.Room{}, ErrTargetOwner
	}
	policy := service.policy(room, action)
	allowed, err := service.hasPermission(ctx, actorID, policy.anyNode)
	if err != nil {
		return roommodel.Room{}, err
	}
	if !allowed {
		allowed, err = service.localAllowed(ctx, room, actorID, policy)
		if err != nil || !allowed {
			return roommodel.Room{}, permissionError(err)
		}
	}
	protected, err := service.hasPermission(ctx, targetID, service.nodes.Unkickable)
	if err != nil {
		return roommodel.Room{}, err
	}
	if protected {
		return roommodel.Room{}, ErrTargetProtected
	}

	return room, nil
}

// authorizeSystemTarget enforces owner and staff immunity for internal room actions.
func (service *Service) authorizeSystemTarget(ctx context.Context, roomID int64, targetID int64) error {
	room, found, err := service.rooms.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if !found {
		return ErrRoomNotFound
	}
	if room.OwnerPlayerID == targetID {
		return ErrTargetOwner
	}
	protected, err := service.hasPermission(ctx, targetID, service.nodes.Unkickable)
	if err != nil {
		return err
	}
	if protected {
		return ErrTargetProtected
	}
	return nil
}

// localAllowed resolves owner or rights-holder authorization.
func (service *Service) localAllowed(ctx context.Context, room roommodel.Room, actorID int64, policy actionPolicy) (bool, error) {
	if room.OwnerPlayerID != actorID {
		if policy.policy < roommodel.ModerationPolicyOwnerAndRights || service.rights == nil {
			return false, nil
		}
		hasRights, err := service.rights.HasRights(ctx, room.ID, actorID)
		if err != nil || !hasRights {
			return false, err
		}
	}

	return service.hasPermission(ctx, actorID, policy.ownNode)
}

// policy maps one action to room settings and permission nodes.
func (service *Service) policy(room roommodel.Room, action moderationmodel.Action) actionPolicy {
	switch action {
	case moderationmodel.ActionKick:
		return actionPolicy{policy: room.ModerationKick, ownNode: service.nodes.OwnKick, anyNode: service.nodes.AnyKick}
	case moderationmodel.ActionMute, moderationmodel.ActionUnmute:
		return actionPolicy{policy: room.ModerationMute, ownNode: service.nodes.OwnMute, anyNode: service.nodes.AnyMute}
	default:
		return actionPolicy{policy: room.ModerationBan, ownNode: service.nodes.OwnBan, anyNode: service.nodes.AnyBan}
	}
}

// hasPermission resolves one optional global permission node.
func (service *Service) hasPermission(ctx context.Context, playerID int64, node permission.Node) (bool, error) {
	if service.permissions == nil || node == "" {
		return false, nil
	}

	return service.permissions.HasPermission(ctx, playerID, node)
}

// permissionError preserves dependency errors and maps denials.
func permissionError(err error) error {
	if err != nil {
		return err
	}

	return ErrAccessDenied
}
