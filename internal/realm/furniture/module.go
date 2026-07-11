// Package furniture contains furniture realm persistence and runtime wiring.
package furniture

import (
	teleport "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport"
	teleportdb "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport/database"
	teleportpair "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport/pair"
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
		NewGranter,
		NewDefinitionGranter,
		teleport.LoadConfig,
		teleportdb.New,
		NewTeleportPairService,
		teleport.NewService,
	),
	fx.Invoke(teleport.Register),
	fx.Invoke(RegisterConnectionHandlers),
)

// NewTeleportPairService creates validated teleport pairing behavior.
func NewTeleportPairService(store *teleportdb.Repository, furniture service.Manager) *teleportpair.Service {
	return teleportpair.NewService(store, furniture)
}

// NewStore creates the furniture persistence store.
func NewStore(pool *postgres.Pool) repository.Store {
	return repository.New(pool)
}

// NewManager exposes the furniture management contract.
func NewManager(furnitureService *service.Service) service.Manager {
	return furnitureService
}

// NewGranter exposes furniture inventory creation behavior.
func NewGranter(furnitureService *service.Service) service.Granter {
	return furnitureService
}

// NewDefinitionGranter exposes furniture definition and creation behavior.
func NewDefinitionGranter(furnitureService *service.Service) service.DefinitionGranter {
	return furnitureService
}
