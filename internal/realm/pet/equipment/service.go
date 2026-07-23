// Package equipment owns pet products, saddle state, and riding.
package equipment

import (
	"context"
	"errors"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	petfed "github.com/niflaot/pixels/internal/realm/pet/care/events/fed"
	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outhanditem "github.com/niflaot/pixels/networking/outbound/room/entities/handitem"
)

// Service coordinates owned furniture products and pet mutations.
type Service struct {
	// config stores normalized lifecycle and product policy.
	config petpolicy.Config
	// store persists pet mutations and transaction scope.
	store petrecord.Store
	// references resolves immutable product rules.
	references petreference.Reader
	// furniture owns furniture inventory state.
	furniture furnitureservice.TradingManager
	// furnitureRooms moves placed product items before transactional consumption.
	furnitureRooms furnitureservice.Manager
	// rooms resolves active room generations.
	rooms *roomlive.Registry
	// runtime owns active pet state and projection.
	runtime *petruntime.Service
	// connections broadcasts unit hand-item changes.
	connections *netconn.Registry
}

// New creates pet equipment behavior.
func New(config petpolicy.Config, store petrecord.Store, references petreference.Reader, furniture furnitureservice.TradingManager, furnitureRooms furnitureservice.Manager, rooms *roomlive.Registry, runtime *petruntime.Service, connections *netconn.Registry) *Service {
	return &Service{config: config.Normalize(), store: store, references: references, furniture: furniture, furnitureRooms: furnitureRooms, rooms: rooms, runtime: runtime, connections: connections}
}

// ProductResult stores one committed product use.
type ProductResult struct {
	// Pet stores the resulting pet state.
	Pet petrecord.Pet
	// Consumed reports whether the furniture item was removed.
	Consumed bool
}

// UseProduct validates, consumes, and projects one typed furniture product.
func (service *Service) UseProduct(ctx context.Context, roomID int64, actorID int64, itemID int64, petID int64) (result ProductResult, err error) {
	started, kind := time.Now(), petobservability.ProductUnknown
	defer func() {
		service.runtime.Metrics().RecordProduct(kind, petobservability.Classify(err, IsExpected(err)))
		service.runtime.Metrics().ObserveTransaction(time.Since(started))
	}()
	pet, found := service.runtime.Snapshot(roomID, petID)
	if !found || pet.OwnerPlayerID != actorID {
		return ProductResult{}, petrecord.ErrNoRights
	}
	active, found := service.rooms.Find(roomID)
	if !found {
		return ProductResult{}, petrecord.ErrInvalidState
	}
	item, found, err := service.furniture.FindItemByID(ctx, itemID)
	if err != nil || !found || item.OwnerPlayerID != actorID || item.RoomID == nil || *item.RoomID != roomID {
		return ProductResult{}, firstError(err, petrecord.ErrInvalidProduct)
	}
	if _, placedFound := active.FurnitureItem(itemID); !placedFound {
		return ProductResult{}, petrecord.ErrInvalidProduct
	}
	rule, err := service.productRule(ctx, item)
	kind = productKind(rule.Kind)
	if err != nil || rule.TypeID >= 0 && rule.TypeID != pet.TypeID {
		return ProductResult{}, firstError(err, petrecord.ErrInvalidProduct)
	}
	if isPlantProduct(rule.Kind) {
		return service.usePlantProduct(ctx, active, actorID, item, pet, rule)
	}
	if !isConsumablePetProduct(rule.Kind) {
		return ProductResult{}, petrecord.ErrInvalidProduct
	}
	if isFood(rule.Kind) && !active.Snapshot().AllowPetsEat {
		return ProductResult{}, petrecord.ErrFeedingDisabled
	}
	before := pet
	result = ProductResult{Consumed: rule.Consumable}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var updated bool
		if rule.Kind == "saddle" {
			result.Pet, updated, err = service.store.SetSaddle(txCtx, pet.ID, actorID, true, pet.Version)
		} else {
			result.Pet, updated, err = service.store.UpdateStats(txCtx, pet.ID, rule.EnergyDelta, rule.HappinessDelta, rule.ExperienceDelta, pet.Version)
		}
		if err != nil || !updated {
			return firstError(err, petrecord.ErrConflict)
		}
		if rule.Consumable {
			if _, pickupErr := service.furnitureRooms.Pickup(txCtx, furnitureservice.PickupParams{ItemID: item.ID, ActorPlayerID: actorID, RoomID: roomID}); pickupErr != nil {
				return pickupErr
			}
			return service.furniture.DeleteInventoryItem(txCtx, item.ID, actorID)
		}
		return nil
	})
	if err != nil {
		return ProductResult{}, err
	}
	service.removeConsumedProduct(ctx, active, item, result.Consumed)
	service.runtime.ReplacePlaced(result.Pet)
	if rule.Kind == "saddle" {
		service.runtime.ProjectFigure(ctx, active, result.Pet)
	} else {
		service.runtime.ProjectStatChange(ctx, active, before, result.Pet, rule.ExperienceDelta, false)
		if isFood(rule.Kind) {
			service.runtime.Publish(ctx, petfed.Name, petfed.Payload{PlayerID: actorID, PetID: pet.ID})
		}
	}
	return result, nil
}

// productKind maps bounded product rule names to telemetry labels.
func productKind(kind string) petobservability.ProductKind {
	switch kind {
	case "food":
		return petobservability.ProductFood
	case "drink":
		return petobservability.ProductDrink
	case "toy":
		return petobservability.ProductToy
	case "nest":
		return petobservability.ProductNest
	case "saddle":
		return petobservability.ProductSaddle
	case "revive", "rebreed", "speed":
		return petobservability.ProductSupplement
	default:
		return petobservability.ProductUnknown
	}
}

// SendProductInventoryChange refreshes the actor's furniture inventory after consumption.
func (service *Service) SendProductInventoryChange(ctx context.Context, target netconn.Context, itemID int64, consumed bool) error {
	if !consumed {
		return nil
	}
	packet, err := outremove.Encode(itemID)
	if err == nil {
		err = target.Send(ctx, packet)
	}
	if err != nil {
		return err
	}
	packet, err = outrefresh.Encode()
	if err != nil {
		return err
	}
	return target.Send(ctx, packet)
}

// productRule resolves one enabled furniture definition rule.
func (service *Service) productRule(ctx context.Context, item furnituremodel.Item) (petrecord.ProductRule, error) {
	references, err := service.references.Current(ctx)
	if err != nil {
		return petrecord.ProductRule{}, err
	}
	rule, found := references.ProductRules[item.DefinitionID]
	if !found || !rule.Enabled {
		return petrecord.ProductRule{}, petrecord.ErrInvalidProduct
	}
	return rule, nil
}

// GiveHandItem consumes the actor's current virtual hand item as food.
func (service *Service) GiveHandItem(ctx context.Context, roomID int64, actorID int64, petUnitID int64) error {
	active, found := service.rooms.Find(roomID)
	if !found || !active.Snapshot().AllowPetsEat {
		return petrecord.ErrFeedingDisabled
	}
	pet, found := service.runtime.FindByUnit(roomID, petUnitID)
	actor, actorFound := active.Unit(actorID)
	if !found || !actorFound || actor.HandItem <= 0 {
		return petrecord.ErrInvalidProduct
	}
	saved, updated, err := service.store.UpdateStats(ctx, pet.ID, 10, 2, 1, pet.Version)
	if err != nil || !updated {
		return firstError(err, petrecord.ErrConflict)
	}
	unit, err := active.SetHandItem(actorID, 0)
	if err == nil {
		if packet, encodeErr := outhanditem.Encode(unit.UnitID, 0); encodeErr == nil {
			_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
		}
	}
	service.runtime.ProjectStatChange(ctx, active, pet, saved, 1, false)
	service.runtime.Publish(ctx, petfed.Name, petfed.Payload{PlayerID: actorID, PetID: pet.ID})
	return nil
}

// isFood reports whether a product requires the room feeding flag.
func isFood(kind string) bool {
	return kind == "food" || kind == "drink" || kind == "toy" || kind == "nest"
}

// isConsumablePetProduct reports whether equipment owns the product behavior.
func isConsumablePetProduct(kind string) bool { return isFood(kind) || kind == "saddle" }

// isPlantProduct reports whether a placed product mutates monsterplant lifecycle.
func isPlantProduct(kind string) bool {
	return kind == "revive" || kind == "rebreed" || kind == "speed"
}

// firstError chooses infrastructure failures over expected domain errors.
func firstError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}

// IsExpected reports errors that must not disconnect Nitro.
func IsExpected(err error) bool {
	return errors.Is(err, petrecord.ErrPetNotFound) || errors.Is(err, petrecord.ErrNoRights) || errors.Is(err, petrecord.ErrInvalidProduct) || errors.Is(err, petrecord.ErrFeedingDisabled) || errors.Is(err, petrecord.ErrInvalidState) || errors.Is(err, petrecord.ErrConflict)
}
