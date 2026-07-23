package service

import (
	"context"
	"sort"

	"github.com/niflaot/pixels/internal/permission"
	permissioncache "github.com/niflaot/pixels/internal/permission/cache"
	permissionchanged "github.com/niflaot/pixels/internal/permission/events/changed"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	permissionrepo "github.com/niflaot/pixels/internal/permission/repository"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// Service implements permission resolution and mutation behavior.
type Service struct {
	// store persists permission records.
	store permissionrepo.Store
	// cache stores shared and process-local permission fragments.
	cache *permissioncache.Cache
	// events publishes committed permission changes.
	events bus.Publisher
	// log records non-critical event delivery failures.
	log *zap.Logger
}

// New creates a permission service.
func New(store permissionrepo.Store, cache *permissioncache.Cache, events bus.Publisher, log *zap.Logger) *Service {
	if cache == nil {
		cache = permissioncache.New(nil, log)
	}
	if log == nil {
		log = zap.NewNop()
	}

	return &Service{store: store, cache: cache, events: events, log: log}
}

// Groups lists all active permission groups.
func (service *Service) Groups(ctx context.Context) ([]permissionmodel.Group, error) {
	return service.store.ListGroups(ctx)
}

// PrimaryGroup returns the player's highest-weight permission group.
func (service *Service) PrimaryGroup(ctx context.Context, playerID int64) (permissionmodel.Group, bool, error) {
	groups, err := service.playerGroups(ctx, playerID)
	if err != nil || len(groups) == 0 {
		return permissionmodel.Group{}, false, err
	}

	return groups[0], true, nil
}

// EffectiveNodes returns registered nodes decided by player or group grants.
func (service *Service) EffectiveNodes(ctx context.Context, playerID int64) ([]ResolvedNode, error) {
	if playerID <= 0 {
		return nil, ErrInvalidPlayerID
	}

	resolved := make([]ResolvedNode, 0)
	for _, node := range permission.AllNodes() {
		decision, err := service.resolve(ctx, playerID, node)
		if err != nil {
			return nil, err
		}
		if decision.found {
			resolved = append(resolved, ResolvedNode{Node: node, Allowed: decision.allowed, Source: decision.source})
		}
	}

	return resolved, nil
}

// EffectivePerks returns allowed Nitro perk names.
func (service *Service) EffectivePerks(ctx context.Context, playerID int64) ([]string, error) {
	perks := make([]string, 0)
	for _, registration := range permission.RegisteredNodes() {
		if registration.PerkName == "" {
			continue
		}
		allowed, err := service.HasPermission(ctx, playerID, registration.Node)
		if err != nil {
			return nil, err
		}
		if allowed {
			perks = append(perks, registration.PerkName)
		}
	}
	sort.Strings(perks)

	return perks, nil
}

// AffectedPlayerIDs lists players inheriting from one group.
func (service *Service) AffectedPlayerIDs(ctx context.Context, groupID int64) ([]int64, error) {
	if groupID <= 0 {
		return nil, ErrInvalidGroupID
	}

	return service.store.ListAffectedPlayerIDs(ctx, groupID)
}

// publish emits one committed permission change without failing its mutation.
func (service *Service) publish(ctx context.Context, payload permissionchanged.Payload) {
	if service.events == nil {
		return
	}
	if err := service.events.Publish(ctx, bus.Event{Name: permissionchanged.Name, Payload: payload}); err != nil {
		service.log.Warn("permission change projection failed", zap.Int64("player_id", payload.PlayerID), zap.Any("group_id", payload.GroupID), zap.Error(err))
	}
}
