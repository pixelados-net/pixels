package service

import (
	"context"
	"strings"

	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
)

// validateGroup validates fields, parent existence, and inheritance acyclicity.
func (service *Service) validateGroup(ctx context.Context, group permissionmodel.Group, groupID int64) error {
	if len(group.Name) < 1 || len(group.Name) > 40 || len(group.Prefix) > 80 || len(group.PrefixColor) > 32 || group.RoomEffectID != nil && *group.RoomEffectID <= 0 {
		return ErrInvalidGroup
	}
	existing, found, err := service.store.FindGroupByName(ctx, group.Name)
	if err != nil {
		return err
	}
	if found && existing.ID != groupID {
		return ErrConflict
	}
	if group.ParentGroupID == nil {
		return nil
	}
	if *group.ParentGroupID <= 0 || *group.ParentGroupID == groupID {
		return ErrInheritanceCycle
	}

	visited := map[int64]struct{}{groupID: {}}
	parentID := group.ParentGroupID
	for parentID != nil {
		if _, exists := visited[*parentID]; exists {
			return ErrInheritanceCycle
		}
		visited[*parentID] = struct{}{}
		parent, found, err := service.store.FindGroupByID(ctx, *parentID)
		if err != nil {
			return err
		}
		if !found {
			return ErrGroupNotFound
		}
		parentID = parent.ParentGroupID
	}

	return nil
}

// validateGroupNode validates a group and grant identifier.
func (service *Service) validateGroupNode(ctx context.Context, groupID int64, node permission.Node) error {
	if groupID <= 0 {
		return ErrInvalidGroupID
	}
	if !validGrantNode(node) {
		return ErrInvalidNode
	}
	_, found, err := service.store.FindGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if !found {
		return ErrGroupNotFound
	}

	return nil
}

// validateMembership validates player and group identities.
func (service *Service) validateMembership(ctx context.Context, playerID int64, groupID int64) error {
	if playerID <= 0 {
		return ErrInvalidPlayerID
	}
	if groupID <= 0 {
		return ErrInvalidGroupID
	}
	_, found, err := service.store.FindGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if !found {
		return ErrGroupNotFound
	}

	return nil
}

// validatePlayerNode validates a player and direct grant identifier.
func validatePlayerNode(playerID int64, node permission.Node) error {
	if playerID <= 0 {
		return ErrInvalidPlayerID
	}
	if !validGrantNode(node) {
		return ErrInvalidNode
	}

	return nil
}

// validGrantNode reports whether a grant is registered or a meaningful wildcard.
func validGrantNode(node permission.Node) bool {
	if !node.Valid() {
		return false
	}
	if node == permission.Node(permission.Wildcard) || permission.Registered(node) {
		return true
	}
	if !strings.HasSuffix(string(node), "."+permission.Wildcard) {
		return false
	}
	for _, registered := range permission.AllNodes() {
		if node.Matches(registered) {
			return true
		}
	}

	return false
}

// applyGroupUpdate applies optional group mutation fields.
func applyGroupUpdate(group *permissionmodel.Group, params UpdateGroupParams) {
	if params.Name != nil {
		group.Name = strings.TrimSpace(*params.Name)
	}
	if params.Weight != nil {
		group.Weight = *params.Weight
	}
	if params.Prefix != nil {
		group.Prefix = strings.TrimSpace(*params.Prefix)
	}
	if params.PrefixColor != nil {
		group.PrefixColor = strings.TrimSpace(*params.PrefixColor)
	}
	if params.RoomEffectID != nil {
		group.RoomEffectID = *params.RoomEffectID
	}
	if params.ParentGroupID != nil {
		group.ParentGroupID = *params.ParentGroupID
	}
}
