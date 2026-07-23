package service

import (
	"context"
	"strings"

	"github.com/niflaot/pixels/internal/permission"
	permissionchanged "github.com/niflaot/pixels/internal/permission/events/changed"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
)

const (
	// defaultGroupName identifies the group assigned to new players.
	defaultGroupName = "member"
)

// CreateGroup creates one permission group.
func (service *Service) CreateGroup(ctx context.Context, params CreateGroupParams) (permissionmodel.Group, error) {
	group := permissionmodel.Group{Name: strings.TrimSpace(params.Name), Weight: params.Weight,
		Prefix: strings.TrimSpace(params.Prefix), PrefixColor: strings.TrimSpace(params.PrefixColor), RoomEffectID: params.RoomEffectID, ParentGroupID: params.ParentGroupID}
	if err := service.validateGroup(ctx, group, 0); err != nil {
		return permissionmodel.Group{}, err
	}

	return service.store.CreateGroup(ctx, group)
}

// UpdateGroup applies a partial permission group mutation.
func (service *Service) UpdateGroup(ctx context.Context, groupID int64, params UpdateGroupParams) (permissionmodel.Group, error) {
	group, found, err := service.store.FindGroupByID(ctx, groupID)
	if err != nil {
		return permissionmodel.Group{}, err
	}
	if !found {
		return permissionmodel.Group{}, ErrGroupNotFound
	}
	applyGroupUpdate(&group, params)
	if err := service.validateGroup(ctx, group, groupID); err != nil {
		return permissionmodel.Group{}, err
	}
	affected, err := service.store.ListAffectedPlayerIDs(ctx, groupID)
	if err != nil {
		return permissionmodel.Group{}, err
	}
	updated, changed, err := service.store.UpdateGroup(ctx, group)
	if err != nil {
		return permissionmodel.Group{}, err
	}
	if !changed {
		return permissionmodel.Group{}, ErrConflict
	}
	service.cache.InvalidateGroup(ctx, groupID)
	for _, playerID := range affected {
		service.cache.InvalidatePlayerGroups(ctx, playerID)
	}
	service.publish(ctx, permissionchanged.Payload{GroupID: &groupID})

	return updated, nil
}

// GrantGroupNode creates or replaces one group grant.
func (service *Service) GrantGroupNode(ctx context.Context, groupID int64, node permission.Node, allowed bool) error {
	if err := service.validateGroupNode(ctx, groupID, node); err != nil {
		return err
	}
	if err := service.store.UpsertGroupNode(ctx, groupID, node, allowed); err != nil {
		return err
	}
	service.cache.InvalidateGroup(ctx, groupID)
	service.publish(ctx, permissionchanged.Payload{GroupID: &groupID})

	return nil
}

// RevokeGroupNode removes one group grant.
func (service *Service) RevokeGroupNode(ctx context.Context, groupID int64, node permission.Node) error {
	if err := service.validateGroupNode(ctx, groupID, node); err != nil {
		return err
	}
	if err := service.store.DeleteGroupNode(ctx, groupID, node); err != nil {
		return err
	}
	service.cache.InvalidateGroup(ctx, groupID)
	service.publish(ctx, permissionchanged.Payload{GroupID: &groupID})

	return nil
}

// AddPlayerToGroup creates one player membership.
func (service *Service) AddPlayerToGroup(ctx context.Context, playerID int64, groupID int64) error {
	if err := service.validateMembership(ctx, playerID, groupID); err != nil {
		return err
	}
	if err := service.store.AddPlayerToGroup(ctx, playerID, groupID); err != nil {
		return err
	}
	service.cache.InvalidatePlayerGroups(ctx, playerID)
	service.publish(ctx, permissionchanged.Payload{PlayerID: playerID})

	return nil
}

// RemovePlayerFromGroup removes one player membership.
func (service *Service) RemovePlayerFromGroup(ctx context.Context, playerID int64, groupID int64) error {
	if err := service.validateMembership(ctx, playerID, groupID); err != nil {
		return err
	}
	if err := service.store.RemovePlayerFromGroup(ctx, playerID, groupID); err != nil {
		return err
	}
	service.cache.InvalidatePlayerGroups(ctx, playerID)
	service.publish(ctx, permissionchanged.Payload{PlayerID: playerID})

	return nil
}

// GrantPlayerNode creates or replaces one direct player grant.
func (service *Service) GrantPlayerNode(ctx context.Context, playerID int64, node permission.Node, allowed bool) error {
	if err := validatePlayerNode(playerID, node); err != nil {
		return err
	}
	if err := service.store.UpsertPlayerNode(ctx, playerID, node, allowed); err != nil {
		return err
	}
	service.cache.InvalidatePlayerNodes(ctx, playerID)
	service.publish(ctx, permissionchanged.Payload{PlayerID: playerID})

	return nil
}

// RevokePlayerNode removes one direct player grant.
func (service *Service) RevokePlayerNode(ctx context.Context, playerID int64, node permission.Node) error {
	if err := validatePlayerNode(playerID, node); err != nil {
		return err
	}
	if err := service.store.DeletePlayerNode(ctx, playerID, node); err != nil {
		return err
	}
	service.cache.InvalidatePlayerNodes(ctx, playerID)
	service.publish(ctx, permissionchanged.Payload{PlayerID: playerID})

	return nil
}

// AssignDefaultGroup adds a player to the member group when it exists.
func (service *Service) AssignDefaultGroup(ctx context.Context, playerID int64) error {
	if playerID <= 0 {
		return ErrInvalidPlayerID
	}
	group, found, err := service.store.FindGroupByName(ctx, defaultGroupName)
	if err != nil {
		return err
	}
	if !found {
		return ErrGroupNotFound
	}

	return service.AddPlayerToGroup(ctx, playerID, group.ID)
}
