// Package furniture contains furniture realm persistence and runtime wiring.
package furniture

import (
	"context"
	"github.com/niflaot/pixels/internal/realm/furniture/interactions"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
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
		NewStateUpdater,
		NewGranter,
		NewDefinitionGranter,
		NewRoomBundleManager,
		NewTradingManager,
		interactions.NewRegistry,
		essential.NewWithEffects,
		teleport.LoadConfig,
		teleportdb.New,
		NewTeleportPairService,
		NewTeleportPairer,
		teleport.NewService,
	),
	fx.Invoke(teleport.Register),
	fx.Invoke(essential.Register),
	fx.Invoke(RegisterConnectionHandlers),
)

// NewTeleportPairService creates validated teleport pairing behavior.
func NewTeleportPairService(store *teleportdb.Repository, furniture service.Manager) *teleportpair.Service {
	return teleportpair.NewService(store, furniture)
}

// NewTradingManager exposes guarded Marketplace and direct-trade ownership mutations.
func NewTradingManager(furnitureService *service.Service) service.TradingManager {
	return furnitureService
}

// teleportPairer adapts validated teleport relationships to purchase workflows.
type teleportPairer struct {
	// pairs manages durable teleport relationships.
	pairs *teleportpair.Service
}

// NewTeleportPairer exposes teleport pairing without leaking pair records.
func NewTeleportPairer(pairs *teleportpair.Service) service.TeleportPairer {
	return teleportPairer{pairs: pairs}
}

// PairTeleports validates and pairs two teleport items owned by one player.
func (pairer teleportPairer) PairTeleports(ctx context.Context, ownerPlayerID int64, firstItemID int64, secondItemID int64) error {
	_, err := pairer.pairs.PairGranted(ctx, ownerPlayerID, firstItemID, secondItemID)

	return err
}

// NewStore creates the furniture persistence store.
func NewStore(pool *postgres.Pool) repository.Store {
	return repository.New(pool)
}

// NewManager exposes the furniture management contract.
func NewManager(furnitureService *service.Service) service.Manager {
	return furnitureService
}

// NewStateUpdater exposes guarded furniture interaction state mutations.
func NewStateUpdater(furnitureService *service.Service) service.StateUpdater {
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

// NewRoomBundleManager exposes efficient room furniture cloning.
func NewRoomBundleManager(furnitureService *service.Service) service.RoomBundleManager {
	return furnitureService
}
