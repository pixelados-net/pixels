package catalog

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	petidentity "github.com/niflaot/pixels/internal/realm/pet/identity"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outrequested "github.com/niflaot/pixels/networking/outbound/room/pet/package/requested"
)

// UseFurniture handles placed package and monsterplant seed interactions.
func (service *Service) UseFurniture(ctx context.Context, request essential.Request) (bool, error) {
	switch request.Item.Definition.InteractionType {
	case "monsterplant_seed":
		return true, service.useSeed(ctx, request, service.runtime.Now())
	case "pet_package":
		return true, service.requestPackageName(ctx, request)
	default:
		return false, nil
	}
}

// requestPackageName opens Nitro's naming prompt for one owned placed package.
func (service *Service) requestPackageName(ctx context.Context, request essential.Request) error {
	if request.Item.OwnerPlayerID != request.PlayerID || request.Target.ConnectionID == "" {
		return petrecord.ErrNoRights
	}
	typeID, err := strconv.ParseInt(strings.TrimSpace(request.Item.Definition.CustomParams), 10, 32)
	if err != nil || typeID < 0 || typeID > 35 {
		return petrecord.ErrInvalidProduct
	}
	references, err := service.references.Current(ctx)
	if err != nil {
		return err
	}
	breed, found := firstSellableBreed(references, int32(typeID))
	if !found {
		return petrecord.ErrInvalidAppearance
	}
	figure := fmt.Sprintf("%d %s %d", typeID, breed.Color, breed.BreedID)
	packet, err := outrequested.Encode(request.Item.ID, figure)
	if err != nil {
		return err
	}
	if err = request.Target.Send(ctx, packet); err != nil {
		return err
	}
	expiresAt := service.runtime.Now().Add(service.config.PackageTimeout)
	service.packageMutex.Lock()
	service.packageRequests[request.Item.ID] = packageRequest{ownerID: request.PlayerID, roomID: request.Room.ID(), expiresAt: expiresAt}
	service.packageMutex.Unlock()
	request.Room.Schedule(service.config.PackageTimeout, func(now time.Time) {
		service.expirePackageRequest(request.Item.ID, request.PlayerID, request.Room.ID(), expiresAt, now)
	})
	return nil
}

// expirePackageRequest removes one unchanged naming prompt at its room-owned deadline.
func (service *Service) expirePackageRequest(itemID int64, ownerID int64, roomID int64, expiresAt time.Time, now time.Time) {
	service.packageMutex.Lock()
	request, found := service.packageRequests[itemID]
	if found && request.ownerID == ownerID && request.roomID == roomID && request.expiresAt.Equal(expiresAt) && !now.Before(expiresAt) {
		delete(service.packageRequests, itemID)
	}
	service.packageMutex.Unlock()
}

// useSeed atomically consumes one placed seed and creates its plant on that tile.
func (service *Service) useSeed(ctx context.Context, request essential.Request, now time.Time) error {
	if request.Item.OwnerPlayerID != request.PlayerID || request.Target.ConnectionID == "" {
		return petrecord.ErrNoRights
	}
	references, err := service.references.Current(ctx)
	if err != nil {
		return err
	}
	item, found, err := service.furniture.FindItemByID(ctx, request.Item.ID)
	if err != nil || !found || item.OwnerPlayerID != request.PlayerID || item.RoomID == nil || *item.RoomID != request.Room.ID() {
		return firstError(err, petrecord.ErrInvalidProduct)
	}
	rule, found := references.ProductRules[item.DefinitionID]
	if !found || !rule.Enabled || rule.Kind != "seed" || rule.TypeID < 0 {
		return petrecord.ErrInvalidProduct
	}
	breed, found := firstEnabledBreed(references, rule.TypeID)
	if !found {
		return petrecord.ErrInvalidAppearance
	}
	_, err = request.Room.ReloadFurniture(request.Item.ID, nil)
	if err != nil {
		return err
	}
	seedEntityKey := petruntime.EntityKey(-request.Item.ID)
	unit, err := request.Room.AddEntity(seedEntityKey, request.PlayerID, worldunit.KindPet, worldpath.Position{Point: request.Item.Point}, request.Item.Rotation)
	if err != nil {
		_, _ = request.Room.ReloadFurniture(request.Item.ID, &request.Item)
		return petrecord.ErrInvalidState
	}
	plant := petrecord.Pet{}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		created, _, grantErr := service.store.Grant(txCtx, petrecord.GrantParams{OwnerPlayerID: request.PlayerID, Name: "Monsterplant", TypeID: rule.TypeID, BreedID: breed.BreedID, PaletteID: breed.PaletteID, Color: breed.Color, Parts: petidentity.MonsterPlantAppearance(uint64(request.Item.ID)), OperationKey: SeedOperationKey(request.Item.ID)})
		if grantErr != nil {
			return grantErr
		}
		placed, placedOK, placeErr := service.store.Place(txCtx, created.ID, request.PlayerID, request.Room.ID(), int(request.Item.Point.X), int(request.Item.Point.Y), unit.Position.Z.Units(), int16(request.Item.Rotation), created.Version)
		if placeErr != nil || !placedOK {
			return firstError(placeErr, petrecord.ErrConflict)
		}
		growAt := now.Add(service.config.PlantGrowDuration)
		dieAt := growAt.Add(service.config.PlantLifeDuration)
		plant, placedOK, placeErr = service.store.UpdateLifecycle(txCtx, placed.ID, request.PlayerID, &growAt, &dieAt, placed.Version)
		if placeErr != nil || !placedOK {
			return firstError(placeErr, petrecord.ErrConflict)
		}
		if _, pickupErr := service.furnitureRooms.Pickup(txCtx, furnitureservice.PickupParams{ItemID: request.Item.ID, ActorPlayerID: request.PlayerID, RoomID: request.Room.ID()}); pickupErr != nil {
			return pickupErr
		}
		return service.furniture.DeleteInventoryItem(txCtx, request.Item.ID, request.PlayerID)
	})
	if err != nil {
		request.Room.RemoveEntity(seedEntityKey)
		_, _ = request.Room.ReloadFurniture(request.Item.ID, &request.Item)
		return err
	}
	request.Room.RemoveEntity(seedEntityKey)
	placedUnit, err := request.Room.AddEntity(petruntime.EntityKey(plant.ID), request.PlayerID, worldunit.KindPet, worldpath.Position{Point: request.Item.Point}, request.Item.Rotation)
	if err != nil {
		return err
	}
	_ = placedUnit
	service.runtime.AddPlaced(ctx, plant)
	service.runtime.ProjectSpawn(ctx, request.Room, plant)
	if packet, encodeErr := outremove.Encode(request.Item.ID, request.PlayerID); encodeErr == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, request.Room, packet, 0)
	}
	if packet, encodeErr := outrefresh.Encode(); encodeErr == nil {
		_ = request.Target.Send(ctx, packet)
	}
	return nil
}

// firstEnabledBreed returns the stable lowest enabled breed, including non-sellable plants.
func firstEnabledBreed(references *petreference.Snapshot, typeID int32) (petrecord.Breed, bool) {
	best := petrecord.Breed{}
	found := false
	for _, breed := range references.Breeds {
		if breed.TypeID != typeID || !breed.Enabled {
			continue
		}
		if !found || breed.BreedID < best.BreedID || breed.BreedID == best.BreedID && breed.PaletteID < best.PaletteID {
			best, found = breed, true
		}
	}
	return best, found
}

// SeedOperationKey returns the idempotency key for one seed furniture item.
func SeedOperationKey(itemID int64) string { return fmt.Sprintf("seed:%d", itemID) }
