// Package plant owns monsterplant lifecycle mutations and rewards.
package plant

import (
	"context"
	"errors"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	petharvested "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/events/harvested"
	planttreated "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/events/treated"
	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/world/items"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outsupplemented "github.com/niflaot/pixels/networking/outbound/room/pet/supplemented"
)

// Service coordinates absolute-deadline monsterplant state.
type Service struct {
	// config stores reward and lifecycle policy.
	config petpolicy.Config
	// store persists pet lifecycle mutations.
	store petrecord.Store
	// references resolves species flags.
	references petreference.Reader
	// rewards grants and places plant result furniture in the same transaction.
	rewards furnitureservice.RoomGranter
	// rooms resolves active room generations.
	rooms *roomlive.Registry
	// runtime owns active pet state.
	runtime *petruntime.Service
	// connections broadcasts supplement state.
	connections *netconn.Registry
}

// New creates monsterplant lifecycle behavior.
func New(config petpolicy.Config, store petrecord.Store, references petreference.Reader, rewards furnitureservice.RoomGranter, rooms *roomlive.Registry, runtime *petruntime.Service, connections *netconn.Registry) *Service {
	return &Service{config: config.Normalize(), store: store, references: references, rewards: rewards, rooms: rooms, runtime: runtime, connections: connections}
}

// Supplement applies one protocol-native lifecycle adjustment.
func (service *Service) Supplement(ctx context.Context, roomID int64, petID int64, actorID int64, kind int32) (err error) {
	defer func() {
		expected := errors.Is(err, petrecord.ErrInvalidState) || errors.Is(err, petrecord.ErrInvalidProduct) || errors.Is(err, petrecord.ErrNoRights) || errors.Is(err, petrecord.ErrConflict)
		service.runtime.Metrics().RecordProduct(petobservability.ProductSupplement, petobservability.Classify(err, expected))
	}()
	pet, species, err := service.ownedPlant(ctx, roomID, petID, actorID)
	if err != nil {
		return err
	}
	now := service.runtime.Now()
	growAt, dieAt := pet.GrowAt, pet.DieAt
	switch kind {
	case 0:
		if pet.DerivePlantState(now, species).Dead || dieAt == nil {
			return petrecord.ErrInvalidState
		}
		dieAt = timePointer(now.Add(7 * 24 * time.Hour))
	case 1:
		if growAt == nil || !now.Before(*growAt) {
			return petrecord.ErrInvalidState
		}
		growAt = timePointer(now)
	default:
		return petrecord.ErrInvalidProduct
	}
	saved, updated, err := service.store.UpdateLifecycle(ctx, pet.ID, actorID, growAt, dieAt, pet.Version)
	if err != nil || !updated {
		return firstError(err, petrecord.ErrConflict)
	}
	service.runtime.ReplacePlaced(saved)
	active, found := service.rooms.Find(roomID)
	if found {
		service.runtime.ProjectFigure(ctx, active, saved)
		if packet, encodeErr := outsupplemented.Encode(saved.ID, actorID, kind); encodeErr == nil {
			_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
		}
	}
	service.runtime.Publish(ctx, planttreated.Name, planttreated.Payload{PlayerID: actorID, PetID: pet.ID})
	return nil
}

// Harvest grants one seed and consumes a mature living plant exactly once.
func (service *Service) Harvest(ctx context.Context, target netconn.Context, roomID int64, petID int64, actorID int64) error {
	return service.consumeHarvest(ctx, target, roomID, petID, actorID)
}

// Compost replaces one dead plant with its compost furniture exactly once.
func (service *Service) Compost(ctx context.Context, target netconn.Context, roomID int64, petID int64, actorID int64) error {
	pet, species, err := service.ownedPlant(ctx, roomID, petID, actorID)
	if err != nil {
		return err
	}
	if !pet.DerivePlantState(service.runtime.Now(), species).Dead || pet.X == nil || pet.Y == nil || pet.Rotation == nil {
		return petrecord.ErrInvalidState
	}
	active, found := service.rooms.Find(roomID)
	if !found {
		return petrecord.ErrInvalidState
	}
	placed, definition, worldItem, err := service.createCompost(ctx, active, pet, actorID)
	if err != nil {
		return err
	}
	service.removePlant(ctx, active, pet)
	if _, err = active.ReloadFurniture(placed.ID, &worldItem); err != nil {
		return err
	}
	packet, err := compostPacket(placed, definition, pet.OwnerName)
	if err != nil {
		return err
	}
	if err = broadcast.RoomPacket(ctx, service.connections, active, packet, 0); err != nil {
		return err
	}
	service.runtime.Publish(ctx, petharvested.Name, petharvested.Payload{PetID: pet.ID, OwnerPlayerID: actorID, State: petrecord.StateComposted})
	return nil
}

// consumeHarvest validates maturity, grants one seed, and removes one plant.
func (service *Service) consumeHarvest(ctx context.Context, target netconn.Context, roomID int64, petID int64, actorID int64) error {
	pet, species, err := service.ownedPlant(ctx, roomID, petID, actorID)
	if err != nil {
		return err
	}
	derived := pet.DerivePlantState(service.runtime.Now(), species)
	if !derived.CanHarvest {
		return petrecord.ErrInvalidState
	}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		consumed, consumeErr := service.store.ConsumePlant(txCtx, pet.ID, actorID, roomID, petrecord.StateHarvested, pet.Version)
		if consumeErr != nil || !consumed {
			return firstError(consumeErr, petrecord.ErrConflict)
		}
		_, grantErr := service.rewards.Grant(txCtx, furnitureservice.GrantParams{DefinitionID: service.config.PlantRewardDefinitionID, OwnerPlayerID: actorID, Quantity: 1})
		return grantErr
	})
	if err != nil {
		return err
	}
	if active, found := service.rooms.Find(roomID); found {
		service.removePlant(ctx, active, pet)
	}
	service.runtime.Publish(ctx, petharvested.Name, petharvested.Payload{PetID: pet.ID, OwnerPlayerID: actorID, State: petrecord.StateHarvested})
	if packet, encodeErr := outrefresh.Encode(); encodeErr == nil {
		if err = target.Send(ctx, packet); err != nil {
			return err
		}
	}
	return nil
}

// createCompost atomically consumes a dead plant and places its RIP furniture.
func (service *Service) createCompost(ctx context.Context, active *roomlive.Room, pet petrecord.Pet, actorID int64) (placed furnituremodel.Item, definition furnituremodel.Definition, worldItem worldfurniture.Item, err error) {
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		consumed, consumeErr := service.store.ConsumePlant(txCtx, pet.ID, actorID, active.ID(), petrecord.StateComposted, pet.Version)
		if consumeErr != nil || !consumed {
			return firstError(consumeErr, petrecord.ErrConflict)
		}
		items, grantErr := service.rewards.Grant(txCtx, furnitureservice.GrantParams{DefinitionID: service.config.PlantCompostDefinitionID, OwnerPlayerID: actorID, Quantity: 1})
		if grantErr != nil {
			return grantErr
		}
		if len(items) != 1 {
			return petrecord.ErrConflict
		}
		worldItem, definition, grantErr = service.compostWorldItem(txCtx, pet, items[0])
		if grantErr != nil {
			return grantErr
		}
		placed, grantErr = service.rewards.Place(txCtx, furnitureservice.PlaceParams{
			ItemID: items[0].ID, ActorPlayerID: actorID, RoomID: active.ID(),
			Placement: furnituremodel.Placement{X: *pet.X, Y: *pet.Y, Z: worldItem.Z.Units(), Rotation: furnituremodel.Rotation(*pet.Rotation)},
		})
		return grantErr
	})
	return placed, definition, worldItem, err
}

// compostWorldItem builds a replacement furniture snapshot from the consumed plant position.
func (service *Service) compostWorldItem(ctx context.Context, pet petrecord.Pet, item furnituremodel.Item) (worldfurniture.Item, furnituremodel.Definition, error) {
	definition, found, err := service.rewards.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found {
		return worldfurniture.Item{}, furnituremodel.Definition{}, firstError(err, furnitureservice.ErrDefinitionNotFound)
	}
	worldDefinition, err := roomfurniture.ToWorldDefinition(definition)
	if err != nil {
		return worldfurniture.Item{}, furnituremodel.Definition{}, err
	}
	point, valid := grid.NewPoint(*pet.X, *pet.Y)
	rotation := furnituremodel.Rotation(*pet.Rotation)
	if !valid || pet.Z == nil || !rotation.Valid() {
		return worldfurniture.Item{}, furnituremodel.Definition{}, petrecord.ErrInvalidState
	}
	return worldfurniture.Item{
		ID: item.ID, OwnerPlayerID: item.OwnerPlayerID, Definition: worldDefinition, Point: point,
		Z: grid.HeightFromUnits(*pet.Z), Rotation: worldunit.Rotation(rotation), ExtraData: item.ExtraData,
	}, definition, nil
}

// removePlant clears one consumed plant from runtime and projects its removal.
func (service *Service) removePlant(ctx context.Context, active *roomlive.Room, pet petrecord.Pet) {
	service.runtime.RemovePlaced(active.ID(), pet.ID)
	if unit, removed := active.RemoveEntity(petruntime.EntityKey(pet.ID)); removed {
		service.runtime.ProjectRemove(ctx, active, unit.UnitID)
	}
}

// ownedPlant validates room visibility, ownership, and species lifecycle.
func (service *Service) ownedPlant(ctx context.Context, roomID int64, petID int64, actorID int64) (petrecord.Pet, petrecord.Species, error) {
	pet, found := service.runtime.Snapshot(roomID, petID)
	if !found || pet.OwnerPlayerID != actorID {
		return petrecord.Pet{}, petrecord.Species{}, petrecord.ErrNoRights
	}
	references, err := service.references.Current(ctx)
	if err != nil || pet.TypeID < 0 || pet.TypeID >= int32(len(references.Species)) || !references.SpeciesPresent[pet.TypeID] || !references.Species[pet.TypeID].Plant {
		return petrecord.Pet{}, petrecord.Species{}, firstError(err, petrecord.ErrInvalidState)
	}
	return pet, references.Species[pet.TypeID], nil
}

// timePointer returns one stable time pointer.
func timePointer(value time.Time) *time.Time { return &value }

// firstError chooses infrastructure failures over expected domain errors.
func firstError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}
