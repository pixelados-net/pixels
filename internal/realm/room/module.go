// Package room contains room realm persistence and runtime wiring.
package room

import (
	"github.com/niflaot/pixels/internal/realm/room/layout"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
)

// Module provides room realm persistence state.
var Module = fx.Module(
	"realm-room",
	fx.Provide(
		NewLayoutStore,
		layout.NewService,
		NewLayoutManager,
	),
)

// NewLayoutStore creates the room layout persistence store.
func NewLayoutStore(pool *postgres.Pool) layout.Store {
	return layout.NewRepository(pool)
}

// NewLayoutManager exposes the room layout management contract.
func NewLayoutManager(service *layout.Service) layout.Manager {
	return service
}
