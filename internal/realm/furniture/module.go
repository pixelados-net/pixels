// Package furniture contains furniture realm persistence and runtime wiring.
package furniture

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/furniture/interactions"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	firework "github.com/niflaot/pixels/internal/realm/furniture/interactions/firework"
	gameinteraction "github.com/niflaot/pixels/internal/realm/furniture/interactions/game"
	lovelock "github.com/niflaot/pixels/internal/realm/furniture/interactions/lovelock"
	mysterybox "github.com/niflaot/pixels/internal/realm/furniture/interactions/mysterybox"
	rentable "github.com/niflaot/pixels/internal/realm/furniture/interactions/rentable"
	roller "github.com/niflaot/pixels/internal/realm/furniture/interactions/roller"
	teleport "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport"
	teleportdb "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport/database"
	teleportpair "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport/pair"
	"github.com/niflaot/pixels/internal/realm/furniture/repository"
	"github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
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
		NewStackHeightUpdater,
		NewGranter,
		NewRoomGranter,
		NewDefinitionGranter,
		NewRoomBundleManager,
		NewTradingManager,
		interactions.NewRegistry,
		essential.NewWithEffects,
		firework.LoadConfig,
		firework.New,
		gameinteraction.New,
		roller.LoadConfig,
		roller.New,
		rentable.LoadConfig,
		NewRentableStore,
		rentable.New,
		NewLovelockStore,
		lovelock.New,
		mysterybox.LoadConfig,
		NewMysteryBoxStore,
		mysterybox.New,
		NewMysteryBoxHandler,
		teleport.LoadConfig,
		teleportdb.New,
		NewTeleportPairService,
		NewTeleportPairer,
		teleport.NewService,
	),
	fx.Invoke(teleport.Register),
	fx.Invoke(essential.Register),
	fx.Invoke(RegisterFirework),
	fx.Invoke(gameinteraction.Register),
	fx.Invoke(mysterybox.RegisterBootstrap),
	fx.Invoke(RegisterRollers),
	fx.Invoke(RegisterConnectionHandlers),
)

// RegisterFirework attaches firework behavior to specialized furniture dispatch.
func RegisterFirework(essentialService *essential.Service, service *firework.Service) {
	essentialService.AddExternal(service)
}

// NewRentableStore creates guarded rentable furniture persistence.
func NewRentableStore(pool *postgres.Pool) rentable.Store { return rentable.NewRepository(pool) }

// NewLovelockStore creates concurrency-safe friend furniture persistence.
func NewLovelockStore(pool *postgres.Pool) lovelock.Store { return lovelock.NewRepository(pool) }

// NewMysteryBoxStore creates durable account key tracking.
func NewMysteryBoxStore(pool *postgres.Pool) mysterybox.Store { return mysterybox.NewRepository(pool) }

// NewMysteryBoxHandler composes reusable packet and bootstrap dependencies.
func NewMysteryBoxHandler(players *playerlive.Registry, bindings *binding.Registry, runtime *roomlive.Registry, connections *netconn.Registry, service *mysterybox.Service) mysterybox.Handler {
	return mysterybox.Handler{Players: players, Bindings: bindings, Runtime: runtime, Connections: connections, Service: service}
}

// RegisterRollers attaches roller cycles and their bounded persistence worker.
func RegisterRollers(lifecycle fx.Lifecycle, rooms *roomlive.Registry, service *roller.Service) {
	rooms.AddCyclePublisher(service.Cycle)
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error { service.Start(); return nil },
		OnStop:  func(context.Context) error { service.Stop(); return nil },
	})
}

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

// NewStackHeightUpdater exposes exact custom stack-height mutations.
func NewStackHeightUpdater(furnitureService *service.Service) service.StackHeightUpdater {
	return furnitureService
}

// NewGranter exposes furniture inventory creation behavior.
func NewGranter(furnitureService *service.Service) service.Granter {
	return furnitureService
}

// NewRoomGranter exposes atomic creation and direct room placement behavior.
func NewRoomGranter(furnitureService *service.Service) service.RoomGranter {
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
