// Package membership owns social-group joins, requests, roles, removal, and favorite state.
package membership

import (
	"context"
	"strings"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	groupconfig "github.com/niflaot/pixels/internal/realm/group/config"
	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	"github.com/niflaot/pixels/pkg/bus"
)

// CatalogSelection identifies one validated group catalog purchase.
type CatalogSelection struct {
	// GroupID identifies the selected active group.
	GroupID int64
	// Forum reports a forum entitlement purchase.
	Forum bool
}

// Service coordinates social-group membership mutations and snapshot invalidation.
type Service struct {
	// config stores roster limits and page bounds.
	config groupconfig.Config
	// store persists group membership state.
	store grouprecord.Store
	// permissions resolves explicit hotel staff overrides.
	permissions permissionservice.Checker
	// cache stores immutable player and group generations.
	cache *groupruntime.Cache
	// projector updates current-room unit state after commit.
	projector *groupruntime.Projector
	// metrics stores bounded process-wide group telemetry.
	metrics *groupobservability.Metrics
	// events publishes committed domain changes.
	events bus.Publisher
}

// New creates social-group membership behavior.
func New(config groupconfig.Config, store grouprecord.Store, permissions permissionservice.Checker, cache *groupruntime.Cache, projector *groupruntime.Projector, metrics *groupobservability.Metrics, events bus.Publisher) *Service {
	return &Service{config: config, store: store, permissions: permissions, cache: cache, projector: projector, metrics: metrics, events: events}
}

// PlayerGroups returns one player's active memberships.
func (service *Service) PlayerGroups(ctx context.Context, playerID int64) ([]grouprecord.PlayerGroup, error) {
	if snapshot, found := service.cache.Player(playerID); found {
		return append([]grouprecord.PlayerGroup(nil), snapshot.Groups...), nil
	}
	groups, err := service.store.PlayerGroups(ctx, playerID)
	if err == nil {
		service.cache.PutPlayer(playerID, groups)
	}
	return groups, err
}

// Status returns active membership, pending request, and favorite state.
func (service *Service) Status(ctx context.Context, playerID int64, groupID int64) (grouprecord.Role, bool, bool, bool, error) {
	member, found, err := service.store.Membership(ctx, groupID, playerID)
	if err != nil {
		return 0, false, false, false, err
	}
	pending, err := service.store.Pending(ctx, groupID, playerID)
	if err != nil {
		return 0, false, false, false, err
	}
	favorite := false
	if snapshot, loaded := service.cache.Player(playerID); loaded {
		favorite = snapshot.FavoriteID == groupID
	} else if groups, listErr := service.PlayerGroups(ctx, playerID); listErr == nil {
		for _, item := range groups {
			if item.Group.ID == groupID {
				favorite = item.Favorite
				break
			}
		}
	}
	return member.Role, found, pending, favorite, nil
}

// Information returns active group metadata with one player's current membership state.
func (service *Service) Information(ctx context.Context, playerID int64, groupID int64) (grouprecord.Group, grouprecord.Role, bool, bool, bool, error) {
	group, found, err := service.store.Group(ctx, groupID, false)
	if err != nil {
		return grouprecord.Group{}, 0, false, false, false, err
	}
	if !found {
		return grouprecord.Group{}, 0, false, false, false, grouprecord.ErrNotFound
	}
	role, member, pending, favorite, err := service.Status(ctx, playerID, groupID)
	return group, role, member, pending, favorite, err
}

// MemberPage returns one validated roster page and viewer management flag.
func (service *Service) MemberPage(ctx context.Context, actorID int64, groupID int64, page int32, query string, level int32) (grouprecord.MemberPage, bool, error) {
	started := time.Now()
	if page < 0 || level < 0 || level > 2 || len(query) > service.config.MaxSearchLength {
		return grouprecord.MemberPage{}, false, grouprecord.ErrInvalid
	}
	actor, found, err := service.store.Membership(ctx, groupID, actorID)
	if err != nil {
		return grouprecord.MemberPage{}, false, err
	}
	canManage := found && actor.Role <= grouprecord.Admin
	if !canManage {
		canManage, err = service.has(ctx, actorID, grouppolicy.MembersManageAny)
		if err != nil {
			return grouprecord.MemberPage{}, false, err
		}
	}
	if level == 2 && !canManage {
		return grouprecord.MemberPage{}, false, grouprecord.ErrForbidden
	}
	result, err := service.store.MemberPage(ctx, groupID, page, int32(service.config.MemberPageSize), strings.TrimSpace(query), level)
	service.metrics.Observe(groupobservability.MemberList, time.Since(started))
	service.record(groupobservability.KindList, err)
	return result, canManage, err
}

// Join joins an open group or creates an exclusive-group request.
func (service *Service) Join(ctx context.Context, playerID int64, groupID int64) (grouprecord.Membership, bool, error) {
	member, pending, changed, err := service.store.Join(ctx, groupID, playerID, service.config.MembershipLimit, service.config.MemberLimit, service.config.PendingLimit)
	if err == nil && changed {
		service.projectChange(ctx, groupID, playerID, "join")
	}
	service.record(groupobservability.KindJoin, err)
	return member, pending, err
}

// Add administratively inserts one member or admin idempotently.
func (service *Service) Add(ctx context.Context, actorID int64, groupID int64, playerID int64, role grouprecord.Role) (grouprecord.Membership, bool, error) {
	if err := service.requireRosterManager(ctx, actorID, groupID); err != nil {
		return grouprecord.Membership{}, false, err
	}
	if role == grouprecord.Admin {
		allowed, err := service.has(ctx, actorID, grouppolicy.RolesManageAny)
		actor, found, memberErr := service.store.Membership(ctx, groupID, actorID)
		if memberErr != nil {
			return grouprecord.Membership{}, false, memberErr
		}
		if err != nil || !allowed && (!found || actor.Role != grouprecord.Owner) {
			return grouprecord.Membership{}, false, grouprecord.ErrForbidden
		}
	}
	member, created, err := service.store.AddMember(ctx, groupID, playerID, role, service.config.MembershipLimit, service.config.MemberLimit)
	if err == nil {
		service.projectChange(ctx, groupID, playerID, "add")
	}
	service.record(groupobservability.KindJoin, err)
	return member, created, err
}

// Accept accepts one pending request after social-role authorization.
func (service *Service) Accept(ctx context.Context, actorID int64, groupID int64, playerID int64) (grouprecord.Membership, error) {
	if err := service.requireRosterManager(ctx, actorID, groupID); err != nil {
		return grouprecord.Membership{}, err
	}
	member, err := service.store.AcceptRequest(ctx, groupID, playerID, service.config.MemberLimit)
	if err == nil {
		service.projectChange(ctx, groupID, playerID, "accept")
	}
	service.record(groupobservability.KindAccept, err)
	return member, err
}

// Decline rejects one pending request after social-role authorization.
func (service *Service) Decline(ctx context.Context, actorID int64, groupID int64, playerID int64) (bool, error) {
	if actorID != playerID {
		if err := service.requireRosterManager(ctx, actorID, groupID); err != nil {
			return false, err
		}
	}
	removed, err := service.store.DeclineRequest(ctx, groupID, playerID)
	if err == nil && removed {
		service.projectChange(ctx, groupID, playerID, "decline")
	}
	service.record(groupobservability.KindDecline, err)
	return removed, err
}

// ApproveAll accepts one configured ordered request batch atomically.
func (service *Service) ApproveAll(ctx context.Context, actorID int64, groupID int64) ([]grouprecord.Membership, error) {
	if err := service.requireRosterManager(ctx, actorID, groupID); err != nil {
		return nil, err
	}
	members, err := service.store.ApproveAll(ctx, groupID, service.config.BulkApproveLimit, service.config.MemberLimit)
	if err == nil {
		for _, member := range members {
			service.projectChange(ctx, groupID, member.PlayerID, "approve_all")
		}
	}
	service.record(groupobservability.KindAccept, err)
	return members, err
}

// ChangeRole promotes or demotes one member while protecting the owner.
func (service *Service) ChangeRole(ctx context.Context, actorID int64, groupID int64, playerID int64, role grouprecord.Role) (grouprecord.Membership, error) {
	actor, found, err := service.store.Membership(ctx, groupID, actorID)
	if err != nil {
		return grouprecord.Membership{}, err
	}
	allowed := found && actor.Role == grouprecord.Owner
	if !allowed {
		allowed, err = service.has(ctx, actorID, grouppolicy.RolesManageAny)
		if err != nil {
			return grouprecord.Membership{}, err
		}
	}
	if !allowed {
		return grouprecord.Membership{}, grouprecord.ErrForbidden
	}
	member, err := service.store.ChangeRole(ctx, groupID, playerID, role)
	if err == nil {
		service.projectChange(ctx, groupID, playerID, "role")
	}
	service.record(groupobservability.KindTransfer, err)
	return member, err
}

// ConfirmRemoval returns the current durable HQ furniture count.
func (service *Service) ConfirmRemoval(ctx context.Context, actorID int64, groupID int64, playerID int64) (int, error) {
	if actorID != playerID {
		if err := service.authorizeRemoval(ctx, actorID, groupID, playerID); err != nil {
			return 0, err
		}
	}
	return service.store.FurnitureCount(ctx, groupID, playerID)
}

// Remove removes membership or a self-owned pending request and returns HQ furniture.
func (service *Service) Remove(ctx context.Context, actorID int64, groupID int64, playerID int64) (int, error) {
	if actorID != playerID {
		if err := service.authorizeRemoval(ctx, actorID, groupID, playerID); err != nil {
			return 0, err
		}
	}
	returned, err := service.store.RemoveMember(ctx, groupID, playerID, service.config.FurnitureCleanupLimit)
	if err == nil {
		service.projectChange(ctx, groupID, playerID, "remove")
		if service.projector != nil {
			_ = service.projector.ReturnFurniture(ctx, returned)
		}
	}
	service.record(groupobservability.KindRemove, err)
	service.metrics.Record(groupobservability.HQFurnitureReturn, groupobservability.KindRemove, metricResult(err))
	return returned.Count(), err
}

// ValidateCatalog validates membership and product-specific role policy.
