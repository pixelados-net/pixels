// Package module wires permission persistence, resolution, cache, and projection.
package module

import (
	permissionbroadcast "github.com/niflaot/pixels/internal/permission/broadcast"
	permissioncache "github.com/niflaot/pixels/internal/permission/cache"
	permissionchanged "github.com/niflaot/pixels/internal/permission/events/changed"
	permissionrepo "github.com/niflaot/pixels/internal/permission/repository"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
)

// Module provides permission persistence and runtime behavior.
var Module = fx.Module(
	"permission",
	fx.Provide(
		NewStore,
		permissioncache.New,
		permissionservice.New,
		NewManager,
		NewChecker,
		NewDefaultAssigner,
		permissionbroadcast.NewProjector,
		permissionbroadcast.New,
	),
	fx.Invoke(RegisterBroadcaster),
)

// NewStore creates permission persistence behavior.
func NewStore(pool *postgres.Pool) permissionrepo.Store {
	return permissionrepo.NewFromPool(pool)
}

// NewManager exposes complete permission behavior.
func NewManager(service *permissionservice.Service) permissionservice.Manager {
	return service
}

// NewChecker exposes permission checks to dependent realms.
func NewChecker(service *permissionservice.Service) permissionservice.Checker {
	return service
}

// NewDefaultAssigner exposes default group assignment.
func NewDefaultAssigner(service *permissionservice.Service) permissionservice.DefaultAssigner {
	return service
}

// RegisterBroadcaster subscribes live permission projections.
func RegisterBroadcaster(subscriber bus.Subscriber, broadcaster *permissionbroadcast.Broadcaster) error {
	_, err := subscriber.Subscribe(permissionchanged.Name, bus.PriorityNormal, broadcaster.Handle)
	return err
}
