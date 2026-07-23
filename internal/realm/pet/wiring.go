package pet

import (
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	petbreeding "github.com/niflaot/pixels/internal/realm/pet/breeding"
	petbreedingcommands "github.com/niflaot/pixels/internal/realm/pet/breeding/commands"
	petplant "github.com/niflaot/pixels/internal/realm/pet/breeding/plant"
	petplantcommands "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/commands"
	petcare "github.com/niflaot/pixels/internal/realm/pet/care"
	petcarecommands "github.com/niflaot/pixels/internal/realm/pet/care/commands"
	petcatalog "github.com/niflaot/pixels/internal/realm/pet/catalog"
	petcatalogcommands "github.com/niflaot/pixels/internal/realm/pet/catalog/commands"
	petequipment "github.com/niflaot/pixels/internal/realm/pet/equipment"
	petequipmentcommands "github.com/niflaot/pixels/internal/realm/pet/equipment/commands"
	petpresence "github.com/niflaot/pixels/internal/realm/pet/presence"
	petpresencecommands "github.com/niflaot/pixels/internal/realm/pet/presence/commands"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/bridge"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"go.uber.org/fx"
)

// RegisterCatalogReward connects typed catalog offers to pet persistence.
func RegisterCatalogReward(catalog *catalogservice.Service, pets *petcatalog.Service) {
	catalog.WithPets(pets)
}

// RegisterPetFurniture connects package clicks to the pet package handshake.
func RegisterPetFurniture(interactions *essential.Service, pets *petcatalog.Service) {
	interactions.AddExternal(pets)
}

// NewSpeechInterceptor exposes the WIRED bridge through the pet runtime boundary.
func NewSpeechInterceptor(speech *bridge.SpeechBridge) petruntime.SpeechInterceptor { return speech }

// CommandDeps contains authenticated command dependencies.
type CommandDeps struct {
	fx.In

	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores authenticated connection bindings.
	Bindings *binding.Registry
}

// NewInventoryHandler creates pet inventory command behavior.
func NewInventoryHandler(deps CommandDeps, service *petpresence.Service, runtime *petruntime.Service) petpresencecommands.InventoryHandler {
	return petpresencecommands.InventoryHandler{Service: service, Runtime: runtime, Players: deps.Players, Bindings: deps.Bindings}
}

// NewPlacementHandler creates pet placement command behavior.
func NewPlacementHandler(deps CommandDeps, service *petpresence.Service) petpresencecommands.PlacementHandler {
	return petpresencecommands.PlacementHandler{Service: service, Players: deps.Players, Bindings: deps.Bindings}
}

// NewRoomHandler creates visible pet command behavior.
func NewRoomHandler(deps CommandDeps, service *petpresence.Service, runtime *petruntime.Service) petpresencecommands.RoomHandler {
	return petpresencecommands.RoomHandler{Service: service, Runtime: runtime, Players: deps.Players, Bindings: deps.Bindings}
}

// NewCareHandler creates pet care command behavior.
func NewCareHandler(deps CommandDeps, service *petcare.Service) petcarecommands.Handler {
	return petcarecommands.Handler{Service: service, Players: deps.Players, Bindings: deps.Bindings}
}

// NewEquipmentHandler creates pet equipment command behavior.
func NewEquipmentHandler(deps CommandDeps, service *petequipment.Service) petequipmentcommands.Handler {
	return petequipmentcommands.Handler{Service: service, Players: deps.Players, Bindings: deps.Bindings}
}

// NewPlantHandler creates monsterplant command behavior.
func NewPlantHandler(deps CommandDeps, service *petplant.Service) petplantcommands.Handler {
	return petplantcommands.Handler{Service: service, Players: deps.Players, Bindings: deps.Bindings}
}

// NewBreedingHandler creates pet breeding command behavior.
func NewBreedingHandler(deps CommandDeps, service *petbreeding.Service) petbreedingcommands.Handler {
	return petbreedingcommands.Handler{Service: service, Players: deps.Players, Bindings: deps.Bindings}
}

// NewCatalogHandler creates pet catalog command behavior.
func NewCatalogHandler(deps CommandDeps, service *petcatalog.Service) petcatalogcommands.Handler {
	return petcatalogcommands.Handler{Service: service, Players: deps.Players, Bindings: deps.Bindings}
}
