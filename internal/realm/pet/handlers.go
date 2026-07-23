package pet

import (
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	petbreedingcommands "github.com/niflaot/pixels/internal/realm/pet/breeding/commands"
	petbreedinghandlers "github.com/niflaot/pixels/internal/realm/pet/breeding/handlers"
	petplantcommands "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/commands"
	petplanthandlers "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/handlers"
	petcarecommands "github.com/niflaot/pixels/internal/realm/pet/care/commands"
	petcarehandlers "github.com/niflaot/pixels/internal/realm/pet/care/handlers"
	petcatalogcommands "github.com/niflaot/pixels/internal/realm/pet/catalog/commands"
	petcataloghandlers "github.com/niflaot/pixels/internal/realm/pet/catalog/handlers"
	petequipmentcommands "github.com/niflaot/pixels/internal/realm/pet/equipment/commands"
	petequipmenthandlers "github.com/niflaot/pixels/internal/realm/pet/equipment/handlers"
	petpresencecommands "github.com/niflaot/pixels/internal/realm/pet/presence/commands"
	petpresencehandlers "github.com/niflaot/pixels/internal/realm/pet/presence/handlers"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// HandlerDeps contains all grouped pet protocol handlers.
type HandlerDeps struct {
	fx.In

	// Connections stores the inbound realm registry.
	Connections *realmconn.Handlers
	// Inventory handles inventory reads.
	Inventory petpresencecommands.InventoryHandler
	// Placement handles place and pickup.
	Placement petpresencecommands.PlacementHandler
	// Room handles info, move, and selection.
	Room petpresencecommands.RoomHandler
	// Care handles respect and training.
	Care petcarecommands.Handler
	// Equipment handles products, riding, and flags.
	Equipment petequipmentcommands.Handler
	// Plants handles monsterplant lifecycle.
	Plants petplantcommands.Handler
	// Breeding handles nest sessions.
	Breeding petbreedingcommands.Handler
	// Catalog handles palettes, names, and packages.
	Catalog petcatalogcommands.Handler
	// Log records command dispatch.
	Log *zap.Logger
}

// RegisterHandlers installs every pet client packet adapter once.
func RegisterHandlers(deps HandlerDeps) {
	if deps.Connections == nil || deps.Connections.Inbound == nil {
		return
	}
	registry := deps.Connections.Inbound
	petpresencehandlers.RegisterInventory(registry, petpresencehandlers.NewInventory(deps.Inventory, deps.Log))
	petpresencehandlers.RegisterPlacement(registry, petpresencehandlers.NewPlace(deps.Placement, deps.Log), petpresencehandlers.NewPickup(deps.Placement, deps.Log))
	petpresencehandlers.RegisterRoom(registry, petpresencehandlers.NewInfo(deps.Room, deps.Log), petpresencehandlers.NewMove(deps.Room, deps.Log), petpresencehandlers.NewSelect(deps.Room, deps.Log))
	petcarehandlers.Register(registry, petcarehandlers.NewRespect(deps.Care, deps.Log), petcarehandlers.NewTraining(deps.Care, deps.Log))
	petequipmenthandlers.Register(registry,
		petequipmenthandlers.NewProduct(deps.Equipment, deps.Log), petequipmenthandlers.NewHandItem(deps.Equipment, deps.Log),
		petequipmenthandlers.NewMount(deps.Equipment, deps.Log), petequipmenthandlers.NewSaddle(deps.Equipment, deps.Log),
		petequipmenthandlers.NewRiding(deps.Equipment, deps.Log), petequipmenthandlers.NewBreeding(deps.Equipment, deps.Log),
	)
	petplanthandlers.Register(registry, petplanthandlers.NewSupplement(deps.Plants, deps.Log), petplanthandlers.NewHarvest(deps.Plants, deps.Log), petplanthandlers.NewCompost(deps.Plants, deps.Log))
	petbreedinghandlers.Register(registry, petbreedinghandlers.NewStart(deps.Breeding, deps.Log), petbreedinghandlers.NewCancel(deps.Breeding, deps.Log), petbreedinghandlers.NewConfirm(deps.Breeding, deps.Log))
	petcataloghandlers.Register(registry, petcataloghandlers.NewBreeds(deps.Catalog, deps.Log), petcataloghandlers.NewNameApproval(deps.Catalog, deps.Log), petcataloghandlers.NewPackage(deps.Catalog, deps.Log))
}
