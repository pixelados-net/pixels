package membership

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
)

// SetFavorite sets or clears exactly one active favorite membership.
func (service *Service) SetFavorite(ctx context.Context, playerID int64, groupID *int64) error {
	if err := service.store.SetFavorite(ctx, playerID, groupID); err != nil {
		service.record(groupobservability.KindFavorite, err)
		return err
	}
	service.refresh(ctx, 0, playerID)
	favoriteID := int64(0)
	if groupID != nil {
		favoriteID = *groupID
	}
	service.publishFavorite(ctx, playerID, favoriteID)
	service.record(groupobservability.KindFavorite, nil)
	return nil
}

func (service *Service) ValidateCatalog(ctx context.Context, playerID int64, groupID int64, forum bool) error {
	member, found, err := service.store.Membership(ctx, groupID, playerID)
	if err != nil {
		return err
	}
	if !found || forum && member.Role != grouprecord.Owner {
		return grouprecord.ErrForbidden
	}
	if forum {
		group, found, err := service.store.Group(ctx, groupID, false)
		if err != nil {
			return err
		}
		if !found {
			return grouprecord.ErrNotFound
		}
		if group.ForumEnabled {
			return grouprecord.ErrConflict
		}
	}
	return nil
}

// CommitCatalog links granted items and activates optional forum entitlement.
func (service *Service) CommitCatalog(ctx context.Context, playerID int64, groupID int64, forum bool, itemIDs []int64) error {
	if err := service.ValidateCatalog(ctx, playerID, groupID, forum); err != nil {
		return err
	}
	if err := service.store.LinkFurniture(ctx, groupID, itemIDs); err != nil {
		return err
	}
	if forum {
		enabled, err := service.store.EnableForum(ctx, groupID)
		if err != nil {
			return err
		}
		if !enabled {
			return grouprecord.ErrConflict
		}
	}
	return nil
}

// ProjectCatalog refreshes committed group, buyer, and linked item generations.
func (service *Service) ProjectCatalog(ctx context.Context, playerID int64, groupID int64, itemIDs []int64) {
	service.cache.PutFurnitureLinks(groupID, itemIDs)
	service.refresh(ctx, groupID, playerID)
}

// Requests returns one bounded administration request page.
func (service *Service) Requests(ctx context.Context, actorID int64, groupID int64, offset int, limit int) ([]grouprecord.Request, error) {
	if offset < 0 || limit < 1 || limit > service.config.PendingLimit {
		return nil, grouprecord.ErrInvalid
	}
	if err := service.requireRosterManager(ctx, actorID, groupID); err != nil {
		return nil, err
	}
	return service.store.Requests(ctx, groupID, offset, limit)
}

// requireRosterManager permits owners, admins, and explicit hotel override.
func (service *Service) requireRosterManager(ctx context.Context, actorID int64, groupID int64) error {
	actor, found, err := service.store.Membership(ctx, groupID, actorID)
	if err != nil {
		return err
	}
	if found && actor.Role <= grouprecord.Admin {
		return nil
	}
	allowed, err := service.has(ctx, actorID, grouppolicy.MembersManageAny)
	if err != nil {
		return err
	}
	if !allowed {
		return grouprecord.ErrForbidden
	}
	return nil
}

// authorizeRemoval enforces actor and target rank ordering.
func (service *Service) authorizeRemoval(ctx context.Context, actorID int64, groupID int64, targetID int64) error {
	target, found, err := service.store.Membership(ctx, groupID, targetID)
	if err != nil {
		return err
	}
	if found && target.Role == grouprecord.Owner {
		return grouprecord.ErrForbidden
	}
	actor, actorFound, err := service.store.Membership(ctx, groupID, actorID)
	if err != nil {
		return err
	}
	if actorFound && actor.Role == grouprecord.Owner {
		return nil
	}
	if actorFound && actor.Role == grouprecord.Admin && found && target.Role == grouprecord.Member {
		return nil
	}
	allowed, err := service.has(ctx, actorID, grouppolicy.MembersManageAny)
	if err != nil {
		return err
	}
	if !allowed {
		return grouprecord.ErrForbidden
	}
	return nil
}

// has resolves one optional hotel permission checker.
func (service *Service) has(ctx context.Context, playerID int64, node permission.Node) (bool, error) {
	if service.permissions == nil {
		return false, nil
	}
	return service.permissions.HasPermission(ctx, playerID, node)
}

// refresh replaces affected immutable group and player generations.
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
			if service.projector != nil {
				_ = service.projector.Favorite(ctx, playerID)
			}
		}
	}
}
