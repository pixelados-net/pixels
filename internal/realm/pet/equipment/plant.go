package equipment

import (
	"context"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	planthealed "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/events/healed"
	planttreated "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/events/treated"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	outroomremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outsupplemented "github.com/niflaot/pixels/networking/outbound/room/pet/supplemented"
)

// usePlantProduct consumes one placed lifecycle product and projects its result.
func (service *Service) usePlantProduct(ctx context.Context, active *roomlive.Room, actorID int64, item furnituremodel.Item, pet petrecord.Pet, rule petrecord.ProductRule) (ProductResult, error) {
	references, err := service.references.Current(ctx)
	if err != nil || pet.TypeID < 0 || pet.TypeID >= int32(len(references.Species)) || !references.SpeciesPresent[pet.TypeID] || !references.Species[pet.TypeID].Plant {
		return ProductResult{}, firstError(err, petrecord.ErrInvalidProduct)
	}
	now := service.runtime.Now()
	result := ProductResult{Consumed: rule.Consumable}
	supplementType := int32(0)
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var updated bool
		switch rule.Kind {
		case "revive":
			if !pet.DerivePlantState(now, references.Species[pet.TypeID]).CanRevive {
				return petrecord.ErrInvalidState
			}
			result.Pet, updated, err = service.store.UpdateLifecycle(txCtx, pet.ID, actorID, pet.GrowAt, timeValue(now.Add(service.config.PlantLifeDuration)), pet.Version)
			supplementType = 2
		case "rebreed":
			state := pet.DerivePlantState(now, references.Species[pet.TypeID])
			if !state.FullyGrown || state.Dead || pet.CanBreed {
				return petrecord.ErrInvalidState
			}
			result.Pet, updated, err = service.store.SetBreedingEligibility(txCtx, pet.ID, true, pet.Version)
			supplementType = 3
		case "speed":
			if pet.GrowAt == nil || !now.Before(*pet.GrowAt) {
				return petrecord.ErrInvalidState
			}
			result.Pet, updated, err = service.store.UpdateLifecycle(txCtx, pet.ID, actorID, timeValue(now), pet.DieAt, pet.Version)
			supplementType = 4
		}
		if err != nil || !updated {
			return firstError(err, petrecord.ErrConflict)
		}
		if !rule.Consumable {
			return nil
		}
		if _, pickupErr := service.furnitureRooms.Pickup(txCtx, furnitureservice.PickupParams{ItemID: item.ID, ActorPlayerID: actorID, RoomID: active.ID()}); pickupErr != nil {
			return pickupErr
		}
		return service.furniture.DeleteInventoryItem(txCtx, item.ID, actorID)
	})
	if err != nil {
		return ProductResult{}, err
	}
	service.runtime.ReplacePlaced(result.Pet)
	service.removeConsumedProduct(ctx, active, item, result.Consumed)
	service.runtime.ProjectFigure(ctx, active, result.Pet)
	if packet, encodeErr := outsupplemented.Encode(result.Pet.ID, actorID, supplementType); encodeErr == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
	if rule.Kind == "revive" {
		service.runtime.Publish(ctx, planthealed.Name, planthealed.Payload{PlayerID: actorID, PetID: pet.ID})
	} else {
		service.runtime.Publish(ctx, planttreated.Name, planttreated.Payload{PlayerID: actorID, PetID: pet.ID})
	}
	return result, nil
}

// removeConsumedProduct removes one committed placed product from room state.
func (service *Service) removeConsumedProduct(ctx context.Context, active *roomlive.Room, item furnituremodel.Item, consumed bool) {
	if !consumed {
		return
	}
	_, _ = active.ReloadFurniture(item.ID, nil)
	if packet, err := outroomremove.Encode(item.ID, item.OwnerPlayerID); err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
}

// timeValue returns one stable deadline pointer.
func timeValue(value time.Time) *time.Time { return &value }
