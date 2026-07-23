package identity

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/group/badge"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
)

// Options returns creator cost and room choices.
func (service *Service) Options(ctx context.Context, playerID int64) (int64, []grouprecord.EligibleRoom, error) {
	rooms, err := service.store.EligibleRooms(ctx, playerID)
	return service.config.CreationCost, rooms, err
}

// BadgeRegistry returns the warmed immutable editor generation.
func (service *Service) BadgeRegistry(ctx context.Context) (*badge.Snapshot, error) {
	if snapshot, found := service.registry.Snapshot(); found {
		return snapshot, nil
	}
	if err := service.registry.Refresh(ctx); err != nil {
		return nil, err
	}
	snapshot, _ := service.registry.Snapshot()
	return snapshot, nil
}

// Group returns active group metadata.
func (service *Service) Group(ctx context.Context, groupID int64, includeDeactivated bool) (grouprecord.Group, bool, error) {
	if snapshot, found := service.cache.Group(groupID); found && (!includeDeactivated || snapshot.Group.ID > 0) {
		return snapshot.Group, true, nil
	}
	group, found, err := service.store.Group(ctx, groupID, includeDeactivated)
	if err == nil && found {
		parts, _ := service.store.BadgeParts(ctx, groupID)
		service.cache.PutGroup(groupruntime.GroupSnapshot{Group: group, BadgeParts: parts})
	}
	return group, found, err
}
