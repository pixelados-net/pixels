package identity

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/niflaot/pixels/internal/permission"
	deactivatedevent "github.com/niflaot/pixels/internal/realm/group/identity/events/deactivated"
	updatedevent "github.com/niflaot/pixels/internal/realm/group/identity/events/updated"
	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
)

// Groups returns one bounded administration page.
func (service *Service) Groups(ctx context.Context, filter grouprecord.GroupFilter) ([]grouprecord.Group, error) {
	if filter.Offset < 0 || filter.Limit < 1 || filter.Limit > 200 || filter.State != nil && !filter.State.Valid() {
		return nil, grouprecord.ErrInvalid
	}
	filter.Query = strings.TrimSpace(filter.Query)
	return service.store.Groups(ctx, filter)
}

// Parts returns normalized badge layers for one active group.
func (service *Service) Parts(ctx context.Context, groupID int64) ([]grouprecord.BadgePart, error) {
	if snapshot, found := service.cache.Group(groupID); found {
		return append([]grouprecord.BadgePart(nil), snapshot.BadgeParts...), nil
	}
	return service.store.BadgeParts(ctx, groupID)
}

// Update applies an owner or staff metadata mutation.
func (service *Service) Update(ctx context.Context, actorID int64, groupID int64, version int64, patch grouprecord.GroupPatch) (grouprecord.Group, error) {
	group, err := service.authorize(ctx, actorID, groupID, grouppolicy.ManageAny)
	if err != nil {
		return grouprecord.Group{}, err
	}
	if patch.Name != nil || patch.Description != nil {
		name, description, normalizeErr := service.normalizeIdentity(valueOr(patch.Name, group.Name), valueOr(patch.Description, group.Description))
		if normalizeErr != nil {
			return grouprecord.Group{}, normalizeErr
		}
		patch.Name, patch.Description = &name, &description
	}
	if patch.State != nil && !patch.State.Valid() {
		return grouprecord.Group{}, grouprecord.ErrInvalid
	}
	if patch.ColorA != nil || patch.ColorB != nil {
		if err = service.validateColors(int32Value(patch.ColorA, group.ColorA), int32Value(patch.ColorB, group.ColorB)); err != nil {
			return grouprecord.Group{}, err
		}
	}
	projectsColors := patch.ColorA != nil || patch.ColorB != nil
	updated, err := service.store.UpdateGroup(ctx, groupID, version, patch)
	if err == nil {
		service.refresh(ctx, groupID, 0)
		if projectsColors && service.projector != nil {
			_ = service.projector.GroupFurnitureColors(ctx, groupID)
		}
	}
	service.metrics.Record(groupobservability.Operations, groupobservability.KindUpdate, identityMetricResult(err))
	if err == nil {
		service.publish(ctx, updatedevent.Name, updatedevent.Payload{GroupID: groupID, Version: updated.Version, Action: "identity"})
	}
	return updated, err
}

// SaveBadge validates and replaces a group's badge.
func (service *Service) SaveBadge(ctx context.Context, actorID int64, groupID int64, version int64, parts []grouprecord.BadgePart) (grouprecord.Group, error) {
	if _, err := service.authorize(ctx, actorID, groupID, grouppolicy.BadgeManageAny); err != nil {
		return grouprecord.Group{}, err
	}
	code, normalized, err := service.badges.Compile(parts)
	if err != nil {
		return grouprecord.Group{}, grouprecord.ErrInvalid
	}
	updated, err := service.store.ReplaceBadge(ctx, groupID, version, code, normalized)
	if err == nil {
		service.refresh(ctx, groupID, 0)
	}
	service.metrics.Record(groupobservability.Operations, groupobservability.KindBadge, identityMetricResult(err))
	if err == nil {
		service.publish(ctx, updatedevent.Name, updatedevent.Payload{GroupID: groupID, Version: updated.Version, Action: "badge"})
	}
	return updated, err
}

// Deactivate soft-deactivates a group and invalidates projections.
func (service *Service) Deactivate(ctx context.Context, actorID int64, groupID int64, version int64) (grouprecord.Group, error) {
	if _, err := service.authorize(ctx, actorID, groupID, grouppolicy.DeleteAny); err != nil {
		return grouprecord.Group{}, err
	}
	group, err := service.store.DeactivateGroup(ctx, groupID, version)
	if err == nil {
		service.cache.DeleteGroup(groupID)
		service.refresh(ctx, 0, actorID)
	}
	service.metrics.Record(groupobservability.Operations, groupobservability.KindDeactivate, identityMetricResult(err))
	if err == nil {
		service.publish(ctx, deactivatedevent.Name, deactivatedevent.Payload{GroupID: groupID, Version: group.Version})
	}
	return group, err
}

// Restore restores retained group data to one eligible headquarters.
func (service *Service) Restore(ctx context.Context, actorID int64, groupID int64, version int64, roomID int64) (grouprecord.Group, error) {
	if allowed, err := service.has(ctx, actorID, grouppolicy.DeleteAny); err != nil || !allowed {
		return grouprecord.Group{}, grouprecord.ErrForbidden
	}
	group, err := service.store.RestoreGroup(ctx, groupID, version, roomID)
	if err == nil {
		service.refresh(ctx, groupID, group.OwnerPlayerID)
	}
	service.metrics.Record(groupobservability.Operations, groupobservability.KindRestore, identityMetricResult(err))
	if err == nil {
		service.publish(ctx, updatedevent.Name, updatedevent.Payload{GroupID: groupID, Version: group.Version, Action: "restore"})
	}
	return group, err
}

// TransferOwner atomically changes the protected owner.
func (service *Service) TransferOwner(ctx context.Context, actorID int64, groupID int64, targetID int64, version int64) (grouprecord.Group, error) {
	if allowed, err := service.has(ctx, actorID, grouppolicy.RolesManageAny); err != nil || !allowed {
		return grouprecord.Group{}, grouprecord.ErrForbidden
	}
	group, err := service.store.TransferOwner(ctx, groupID, targetID, version)
	if err == nil {
		service.refresh(ctx, groupID, targetID)
	}
	service.metrics.Record(groupobservability.Operations, groupobservability.KindTransfer, identityMetricResult(err))
	if err == nil {
		service.publish(ctx, updatedevent.Name, updatedevent.Payload{GroupID: groupID, Version: group.Version, Action: "owner"})
	}
	return group, err
}

// RebindRoom atomically changes a group headquarters.
func (service *Service) RebindRoom(ctx context.Context, actorID int64, groupID int64, roomID int64, version int64) (grouprecord.Group, error) {
	if allowed, err := service.has(ctx, actorID, grouppolicy.HomeRoomRebind); err != nil || !allowed {
		return grouprecord.Group{}, grouprecord.ErrForbidden
	}
	group, err := service.store.RebindRoom(ctx, groupID, roomID, version)
	if err == nil {
		service.refresh(ctx, groupID, group.OwnerPlayerID)
	}
	service.metrics.Record(groupobservability.Operations, groupobservability.KindRebind, identityMetricResult(err))
	if err == nil {
		service.publish(ctx, updatedevent.Name, updatedevent.Payload{GroupID: groupID, Version: group.Version, Action: "home_room"})
	}
	return group, err
}

// authorize permits the owner or one explicit hotel override.
func (service *Service) authorize(ctx context.Context, actorID int64, groupID int64, node permission.Node) (grouprecord.Group, error) {
	group, found, err := service.store.Group(ctx, groupID, false)
	if err != nil || !found {
		return grouprecord.Group{}, grouprecord.ErrNotFound
	}
	if group.OwnerPlayerID == actorID {
		return group, nil
	}
	allowed, err := service.has(ctx, actorID, node)
	if err != nil || !allowed {
		return grouprecord.Group{}, grouprecord.ErrForbidden
	}
	return group, nil
}

// has resolves one optional permission checker.
func (service *Service) has(ctx context.Context, actorID int64, node permission.Node) (bool, error) {
	if service.permissions == nil {
		return false, nil
	}
	return service.permissions.HasPermission(ctx, actorID, node)
}

// normalizeIdentity trims, bounds, and filters hotel-facing identity text.
func (service *Service) normalizeIdentity(name string, description string) (string, string, error) {
	name, description = strings.TrimSpace(name), strings.TrimSpace(description)
	if utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 29 || utf8.RuneCountInString(description) > 254 {
		return "", "", grouprecord.ErrInvalid
	}
	if service.filter != nil {
		name, _ = service.filter.Censor(name)
		description, _ = service.filter.Censor(description)
	}
	return name, description, nil
}

// validateColors validates primary and secondary editor color families.
func (service *Service) validateColors(colorA int32, colorB int32) error {
	snapshot, found := service.registry.Snapshot()
	if !found {
		return grouprecord.ErrInvalid
	}
	if _, found = snapshot.Color(grouprecord.BackgroundColor, colorA); !found {
		return grouprecord.ErrInvalid
	}
	if _, found = snapshot.Color(grouprecord.BackgroundColor, colorB); !found {
		return grouprecord.ErrInvalid
	}
	return nil
}

// refresh updates immutable generations after a committed mutation.
func (service *Service) refresh(ctx context.Context, groupID int64, playerID int64) {
	if groupID > 0 {
		if group, found, _ := service.store.Group(ctx, groupID, false); found {
			parts, _ := service.store.BadgeParts(ctx, groupID)
			service.cache.PutGroup(groupruntime.GroupSnapshot{Group: group, BadgeParts: parts})
		}
	}
	if playerID > 0 {
		if groups, err := service.store.PlayerGroups(ctx, playerID); err == nil {
			service.cache.PutPlayer(playerID, groups)
		}
	}
}

// valueOr returns the supplied string pointer or fallback.
func valueOr(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}

// int32Value returns the supplied integer pointer or fallback.
func int32Value(value *int32, fallback int32) int32 {
	if value == nil {
		return fallback
	}
	return *value
}

// ErrInsufficientBalance identifies a failed atomic creation charge.
var ErrInsufficientBalance = errors.New("insufficient balance for social group")
