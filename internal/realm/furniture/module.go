// Package furniture contains furniture realm persistence and runtime wiring.
package furniture

import (
	"github.com/niflaot/pixels/internal/realm/furniture/repository"
	"github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
)

// Module provides furniture realm persistence state.
var Module = fx.Module(
	"realm-furniture",
	fx.Provide(
		NewStore,
		service.New,
		NewManager,
	),
)

// NewStore creates the furniture persistence store.
func NewStore(pool *postgres.Pool) repository.Store {
	return repository.New(pool)
}

// NewManager exposes the furniture management contract.
func NewManager(furnitureService *service.Service) service.Manager {
	return furnitureService
}
