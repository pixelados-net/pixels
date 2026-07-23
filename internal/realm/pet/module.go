// Package pet wires pet persistence, protocol, behavior, and room runtime.
package pet

import (
	petadmin "github.com/niflaot/pixels/internal/realm/pet/admin"
	petbehavior "github.com/niflaot/pixels/internal/realm/pet/behavior"
	petbreeding "github.com/niflaot/pixels/internal/realm/pet/breeding"
	petplant "github.com/niflaot/pixels/internal/realm/pet/breeding/plant"
	petcare "github.com/niflaot/pixels/internal/realm/pet/care"
	petcatalog "github.com/niflaot/pixels/internal/realm/pet/catalog"
	petdb "github.com/niflaot/pixels/internal/realm/pet/database"
	petequipment "github.com/niflaot/pixels/internal/realm/pet/equipment"
	petinventory "github.com/niflaot/pixels/internal/realm/pet/inventory"
	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petpresence "github.com/niflaot/pixels/internal/realm/pet/presence"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	"go.uber.org/fx"
)

// Module provides the complete pet realm.
var Module = fx.Module(
	"realm-pet",
	fx.Provide(
		petpolicy.LoadConfig,
		petdb.New,
		NewStore,
		petreference.New,
		NewReferenceReader,
		petobservability.New,
		petinventory.New,
		petruntime.New,
		petpresence.New,
		petcare.New,
		petbehavior.NewRegistry,
		NewSpeechInterceptor,
		petbehavior.New,
		petequipment.New,
		petplant.New,
		petbreeding.New,
		petcatalog.New,
		petadmin.New,
		NewInventoryHandler,
		NewPlacementHandler,
		NewRoomHandler,
		NewCareHandler,
		NewEquipmentHandler,
		NewPlantHandler,
		NewBreedingHandler,
		NewCatalogHandler,
	),
	fx.Invoke(RegisterHandlers, RegisterRuntime, RegisterCatalogReward, RegisterPetFurniture),
)

// NewStore exposes PostgreSQL persistence through the pet domain contract.
func NewStore(repository *petdb.Repository) petrecord.Store { return repository }

// NewReferenceReader exposes immutable pet references through their domain boundary.
func NewReferenceReader(service *petreference.Service) petreference.Reader { return service }
