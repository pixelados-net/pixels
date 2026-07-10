// Package catalog contains catalog realm dependency wiring.
package catalog

import (
	"context"

	catalogrepo "github.com/niflaot/pixels/internal/realm/catalog/repository"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
)

// Module provides catalog persistence, cache, and purchase behavior.
var Module = fx.Module(
	"realm-catalog",
	fx.Provide(
		NewStore,
		catalogservice.New,
		NewManager,
		NewReader,
	),
	fx.Invoke(RegisterLifecycle),
)

// NewStore creates catalog persistence behavior.
func NewStore(pool *postgres.Pool) catalogrepo.Store {
	return catalogrepo.New(pool)
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
