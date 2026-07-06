// Package room contains room realm persistence and runtime wiring.
package room

import (
	"github.com/niflaot/pixels/internal/realm/room/layout"
	"github.com/niflaot/pixels/internal/realm/room/repository"
	"github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
)

// Module provides room realm persistence state.
var Module = fx.Module(
	"realm-room",
	fx.Provide(
		NewLayoutStore,
		NewStore,
		layout.NewService,
		service.New,
		NewLiveRegistry,
		NewLayoutManager,
		NewManager,
	),
	fx.Invoke(RegisterRuntimeCleanup),
	fx.Invoke(RegisterConnectionHandlers),
)

// NewLayoutStore creates the room layout persistence store.
func NewLayoutStore(pool *postgres.Pool) layout.Store {
	return layout.NewRepository(pool)
}

// NewStore creates the room persistence store.
func NewStore(pool *postgres.Pool) repository.Store {
	return repository.New(pool)
}

// NewLayoutManager exposes the room layout management contract.
func NewLayoutManager(service *layout.Service) layout.Manager {
	return service
}

// NewManager exposes the room management contract.
func NewManager(roomService *service.Service) service.Manager {
	return roomService
}
