// Package catalog contains catalog realm dependency wiring.
package catalog

import (
	"context"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	catalogadmin "github.com/niflaot/pixels/internal/realm/catalog/admin"
	catalogrepo "github.com/niflaot/pixels/internal/realm/catalog/repository"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides catalog persistence, cache, and purchase behavior.
var Module = fx.Module(
	"realm-catalog",
	fx.Provide(
		NewStore,
		NewService,
		NewManager,
		NewReader,
		catalogadmin.New,
		NewAdminManager,
	),
	fx.Invoke(RegisterLifecycle, RegisterConnectionHandlers),
)

// NewService creates permission-aware catalog behavior.
func NewService(store catalogrepo.Store, currencies currencyservice.Granter, furniture furnitureservice.DefinitionGranter, events bus.Publisher, log *zap.Logger, permissions permissionservice.Checker) *catalogservice.Service {
	return catalogservice.New(store, currencies, furniture, events, log, permissions)
}

// NewStore creates catalog persistence behavior.
func NewStore(pool *postgres.Pool) catalogrepo.Store {
	return catalogrepo.New(pool)
}

// NewAdminManager exposes catalog administration behavior.
func NewAdminManager(service *catalogadmin.Service) catalogadmin.Manager {
	return service
}

// NewManager exposes catalog management behavior.
func NewManager(service *catalogservice.Service) catalogservice.Manager {
	return service
}

// NewReader exposes catalog read behavior.
func NewReader(service *catalogservice.Service) catalogservice.Reader {
	return service
}

// RegisterLifecycle loads catalog data when the application starts.
func RegisterLifecycle(lifecycle fx.Lifecycle, service *catalogservice.Service) {
	lifecycle.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		return service.Refresh(ctx)
	}})
}
