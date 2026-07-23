package catalog

import (
	"context"
	"time"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	petidentity "github.com/niflaot/pixels/internal/realm/pet/identity"
	petcreated "github.com/niflaot/pixels/internal/realm/pet/identity/events/created"
	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outreceived "github.com/niflaot/pixels/networking/outbound/inventory/pet/received"
	outremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outresult "github.com/niflaot/pixels/networking/outbound/room/pet/package/result"
)

// OpenPackage consumes one owned furniture package and grants its configured pet once.
func (service *Service) OpenPackage(ctx context.Context, target netconn.Context, ownerID int64, roomID int64, itemID int64, proposedName string) (err error) {
	started, result := time.Now(), petobservability.ResultSuccess
	var metrics *petobservability.Metrics
	if service.runtime != nil {
		metrics = service.runtime.Metrics()
	}
	defer func() {
		if err != nil {
			result = petobservability.Classify(err, result == petobservability.ResultRejected)
		}
		metrics.RecordOperation(petobservability.OperationPackage, result)
		metrics.ObserveTransaction(time.Since(started))
	}()
	name, code := service.ValidateName(proposedName)
	if code != petidentity.NameApproved {
		result = petobservability.ResultRejected
		return service.sendPackageResult(ctx, target, itemID, code, name)
	}
	item, found, err := service.furniture.FindItemByID(ctx, itemID)
	if err == nil && !found {
		granted, operationFound, operationErr := service.store.FindByOperation(ctx, PackageOperationKey(itemID))
		if operationErr != nil {
			return operationErr
		}
		if operationFound && granted.OwnerPlayerID == ownerID {
			return service.sendPackageResult(ctx, target, itemID, petidentity.NameApproved, granted.Name)
		}
	}
	if err != nil || !found || item.OwnerPlayerID != ownerID || item.RoomID == nil || *item.RoomID != roomID {
		if err == nil {
			result = petobservability.ResultRejected
		}
		return firstError(err, service.sendPackageResult(ctx, target, itemID, petidentity.NameInvalidCharacters, ""))
	}
	if !service.validPackageRequest(itemID, ownerID, roomID, service.runtime.Now()) {
		result = petobservability.ResultRejected
		return petrecord.ErrInvalidState
	}
	references, err := service.references.Current(ctx)
	if err != nil {
		return err
	}
	service.clearPackageRequest(itemID)
	rule, found := references.ProductRules[item.DefinitionID]
	if !found || !rule.Enabled || rule.Kind != "package" || rule.TypeID < 0 {
		result = petobservability.ResultRejected
		return petrecord.ErrInvalidProduct
	}
	breed, found := firstSellableBreed(references, rule.TypeID)
	if !found {
		result = petobservability.ResultRejected
		return petrecord.ErrInvalidAppearance
	}
	granted := petrecord.Pet{}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var created bool
		granted, created, err = service.Grant(txCtx, ownerID, rule.TypeID, breed.BreedID, breed.PaletteID, breed.Color, name, PackageOperationKey(itemID))
		if err != nil {
			return err
		}
		if created {
			if _, pickupErr := service.furnitureRooms.Pickup(txCtx, furnitureservice.PickupParams{ItemID: itemID, ActorPlayerID: ownerID, RoomID: roomID}); pickupErr != nil {
				return pickupErr
			}
			return service.furniture.DeleteInventoryItem(txCtx, itemID, ownerID)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if active, activeFound := service.rooms.Find(roomID); activeFound {
		_, _ = active.ReloadFurniture(itemID, nil)
		if packet, encodeErr := outremove.Encode(itemID, ownerID); encodeErr == nil {
			_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
		}
	}
	service.runtime.SendInventoryAdd(ctx, ownerID, granted)
	service.runtime.Publish(ctx, petcreated.Name, petcreated.Payload{PetID: granted.ID, OwnerPlayerID: granted.OwnerPlayerID, TypeID: granted.TypeID})
	if packet, encodeErr := outreceived.Encode(false, petruntime.InventoryPet(granted)); encodeErr == nil {
		_ = target.Send(ctx, packet)
	}
	if packet, encodeErr := outrefresh.Encode(); encodeErr == nil {
		_ = target.Send(ctx, packet)
	}
	return service.sendPackageResult(ctx, target, itemID, petidentity.NameApproved, name)
}

// validPackageRequest reports whether one naming prompt is current.
func (service *Service) validPackageRequest(itemID int64, ownerID int64, roomID int64, now time.Time) bool {
	service.packageMutex.Lock()
	request, found := service.packageRequests[itemID]
	service.packageMutex.Unlock()
	return found && request.ownerID == ownerID && request.roomID == roomID && now.Before(request.expiresAt)
}

// clearPackageRequest removes one completed package naming prompt.
func (service *Service) clearPackageRequest(itemID int64) {
	service.packageMutex.Lock()
	delete(service.packageRequests, itemID)
	service.packageMutex.Unlock()
}

// sendPackageResult sends one native package outcome.
func (service *Service) sendPackageResult(ctx context.Context, target netconn.Context, itemID int64, code int32, info string) error {
	packet, err := outresult.Encode(itemID, code, info)
	if err != nil {
		return err
	}
	return target.Send(ctx, packet)
}

// firstSellableBreed returns the stable lowest composite breed option.
func firstSellableBreed(references *petreference.Snapshot, typeID int32) (petrecord.Breed, bool) {
	best := petrecord.Breed{}
	found := false
	for _, breed := range references.Breeds {
		if breed.TypeID != typeID || !breed.Enabled || !breed.Sellable {
			continue
		}
		if !found || breed.BreedID < best.BreedID || breed.BreedID == best.BreedID && breed.PaletteID < best.PaletteID {
			best, found = breed, true
		}
	}
	return best, found
}
