package equipment

import (
	"context"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	petfed "github.com/niflaot/pixels/internal/realm/pet/care/events/fed"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
)

// ConsumeNeed revalidates and applies one autonomous food, drink, toy, or nest.
func (service *Service) ConsumeNeed(ctx context.Context, roomID int64, petID int64, itemID int64) error {
	active, activeFound := service.rooms.Find(roomID)
	pet, petFound := service.runtime.Snapshot(roomID, petID)
	if !activeFound || !petFound || !active.Snapshot().AllowPetsEat {
		return petrecord.ErrInvalidState
	}
	item, itemFound, err := service.furniture.FindItemByID(ctx, itemID)
	if err != nil || !itemFound || item.RoomID == nil || *item.RoomID != roomID {
		return firstError(err, petrecord.ErrInvalidProduct)
	}
	placed, placedFound := active.FurnitureItem(itemID)
	unit, unitFound := active.UnitMotion(petruntime.EntityKey(petID))
	if !placedFound || !unitFound || unit.Moving || unit.Position.Point != placed.Point {
		return petrecord.ErrInvalidProduct
	}
	rule, err := service.productRule(ctx, item)
	if err != nil || !isFood(rule.Kind) || rule.TypeID >= 0 && rule.TypeID != pet.TypeID {
		return firstError(err, petrecord.ErrInvalidProduct)
	}
	before := pet
	result := ProductResult{Consumed: rule.Consumable}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var updated bool
		result.Pet, updated, err = service.store.UpdateStats(txCtx, pet.ID, rule.EnergyDelta, rule.HappinessDelta, rule.ExperienceDelta, pet.Version)
		if err != nil || !updated {
			return firstError(err, petrecord.ErrConflict)
		}
		if !rule.Consumable {
			return nil
		}
		if _, pickupErr := service.furnitureRooms.Pickup(txCtx, furnitureservice.PickupParams{ItemID: item.ID, ActorPlayerID: item.OwnerPlayerID, RoomID: roomID}); pickupErr != nil {
			return pickupErr
		}
		return service.furniture.DeleteInventoryItem(txCtx, item.ID, item.OwnerPlayerID)
	})
	if err != nil {
		return err
	}
	service.removeConsumedProduct(ctx, active, item, result.Consumed)
	service.runtime.ReplacePlaced(result.Pet)
	service.runtime.ProjectStatChange(ctx, active, before, result.Pet, rule.ExperienceDelta, false)
	service.runtime.Publish(ctx, petfed.Name, petfed.Payload{PlayerID: result.Pet.OwnerPlayerID, PetID: result.Pet.ID})
	return nil
}
