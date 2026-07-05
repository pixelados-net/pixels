// Package navigator contains navigator realm persistence and runtime wiring.
package navigator

import (
	"github.com/niflaot/pixels/internal/realm/navigator/repository"
	"github.com/niflaot/pixels/internal/realm/navigator/service"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
)

// Module provides navigator realm persistence state.
var Module = fx.Module(
	"realm-navigator",
	fx.Provide(
		NewStore,
		service.New,
		NewManager,
	),
)

// NewStore creates the navigator persistence store.
func NewStore(pool *postgres.Pool) repository.Store {
	return repository.New(pool)
}

// NewManager exposes the navigator management contract.
func NewManager(service *service.Service) service.Manager {
	return service
}
