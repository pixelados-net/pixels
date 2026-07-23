// Package presence owns pet inventory and room placement workflows.
package presence

import (
	"context"
	"errors"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	petinventory "github.com/niflaot/pixels/internal/realm/pet/inventory"
	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petpickedup "github.com/niflaot/pixels/internal/realm/pet/presence/events/pickedup"
	petplaced "github.com/niflaot/pixels/internal/realm/pet/presence/events/placed"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// Service coordinates durable and active pet presence.
type Service struct {
	// config stores normalized capacity rules.
	config petpolicy.Config
	// store persists pet aggregates.
	store petrecord.Store
	// rooms resolves active room generations.
	rooms *roomlive.Registry
	// permissions resolves administrative bypasses.
	permissions permissionservice.Checker
	// runtime owns pet room controllers and projections.
	runtime *petruntime.Service
	// inventory caches immutable owner lists.
	inventory *petinventory.Service
}

// New creates pet presence behavior.
func New(config petpolicy.Config, store petrecord.Store, rooms *roomlive.Registry, permissions permissionservice.Checker, runtime *petruntime.Service, inventory *petinventory.Service) *Service {
	return &Service{config: config.Normalize(), store: store, rooms: rooms, permissions: permissions, runtime: runtime, inventory: inventory}
}

// Place compare-and-swaps an owned pet into an active room.
func (service *Service) Place(ctx context.Context, params PlaceParams) (pet petrecord.Pet, err error) {
	started := time.Now()
	pet, err = service.place(ctx, params, false, nil)
	service.runtime.Metrics().RecordOperation(petobservability.OperationPlace, petobservability.Classify(err, IsExpected(err)))
	service.runtime.Metrics().ObserveTransaction(time.Since(started))
	return pet, err
}

// PlaceAdmin places one pet while bypassing owner and room-right checks.
func (service *Service) PlaceAdmin(ctx context.Context, petID int64, roomID int64, point grid.Point, hook TransitionHook) (petrecord.Pet, error) {
	pet, found, err := service.store.Find(ctx, petID)
	if err != nil || !found {
		return petrecord.Pet{}, firstError(err, petrecord.ErrPetNotFound)
	}
	return service.place(ctx, PlaceParams{PetID: petID, ActorPlayerID: pet.OwnerPlayerID, RoomID: roomID, Point: point}, true, hook)
}

// place compare-and-swaps an owned pet with an optional administration bypass.
func (service *Service) place(ctx context.Context, params PlaceParams, bypass bool, hook TransitionHook) (petrecord.Pet, error) {
	if !service.config.Enabled {
		return petrecord.Pet{}, petrecord.ErrPetsDisabled
	}
	pet, found, err := service.store.Find(ctx, params.PetID)
	if err != nil || !found || pet.OwnerPlayerID != params.ActorPlayerID || !pet.Inventory() {
		return petrecord.Pet{}, firstError(err, petrecord.ErrPetNotFound)
	}
	active, found := service.rooms.Find(params.RoomID)
	if !found {
		return petrecord.Pet{}, petrecord.ErrInvalidState
	}
	if err = service.runtime.EnsureRoom(ctx, active); err != nil {
		return petrecord.Pet{}, err
	}
	if !active.Snapshot().AllowPets {
		return petrecord.Pet{}, petrecord.ErrPetsDisabled
	}
	if !bypass {
		allowed, permissionErr := service.canPlace(ctx, active, params.ActorPlayerID)
		if permissionErr != nil || !allowed {
			return petrecord.Pet{}, firstError(permissionErr, petrecord.ErrNoRights)
		}
	}
	if !bypass {
		err = service.checkRoomLimits(ctx, params.RoomID, params.ActorPlayerID)
	}
	if err != nil {
		return petrecord.Pet{}, err
	}
	position := worldpath.Position{Point: params.Point}
	unit, err := active.AddEntity(petruntime.EntityKey(pet.ID), pet.OwnerPlayerID, worldunit.KindPet, position, worldunit.RotationSouth)
	if err != nil {
		return petrecord.Pet{}, petrecord.ErrTileNotFree
	}
	placed := petrecord.Pet{}
	found = false
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var placeErr error
		placed, found, placeErr = service.store.Place(txCtx, pet.ID, pet.OwnerPlayerID, params.RoomID, int(params.Point.X), int(params.Point.Y), unit.Position.Z.Units(), int16(unit.BodyRotation), pet.Version)
		if placeErr != nil || !found || hook == nil {
			return placeErr
		}
		return hook(txCtx, placed)
	})
	if err != nil || !found {
		active.RemoveEntity(petruntime.EntityKey(pet.ID))
		return petrecord.Pet{}, firstError(err, petrecord.ErrConflict)
	}
	service.runtime.AddPlaced(ctx, placed)
	service.runtime.ProjectSpawn(ctx, active, placed)
	service.runtime.SendInventoryRemove(ctx, pet.OwnerPlayerID, pet.ID)
	service.runtime.Publish(ctx, petplaced.Name, petplaced.Payload{PetID: placed.ID, RoomID: params.RoomID, ActorPlayerID: params.ActorPlayerID})
	return placed, nil
}

// Pickup compare-and-swaps a placed pet back into its owner's inventory.
func (service *Service) Pickup(ctx context.Context, params PickupParams) (pet petrecord.Pet, err error) {
	started := time.Now()
	pet, err = service.pickup(ctx, params, false, nil)
	service.runtime.Metrics().RecordOperation(petobservability.OperationPickup, petobservability.Classify(err, IsExpected(err)))
	service.runtime.Metrics().ObserveTransaction(time.Since(started))
	return pet, err
}

// PickupAdmin returns one placed pet without owner or inventory-limit checks.
func (service *Service) PickupAdmin(ctx context.Context, petID int64, roomID int64, hook TransitionHook) (petrecord.Pet, error) {
	pet, found, err := service.store.Find(ctx, petID)
	if err != nil || !found {
		return petrecord.Pet{}, firstError(err, petrecord.ErrPetNotFound)
	}
	return service.pickup(ctx, PickupParams{PetID: petID, ActorPlayerID: pet.OwnerPlayerID, RoomID: roomID}, true, hook)
}

// pickup compare-and-swaps a placed pet with an optional administration bypass.
func (service *Service) pickup(ctx context.Context, params PickupParams, bypass bool, hook TransitionHook) (petrecord.Pet, error) {
	pet, found, err := service.store.Find(ctx, params.PetID)
	if err != nil || !found || pet.RoomID == nil || *pet.RoomID != params.RoomID {
		return petrecord.Pet{}, firstError(err, petrecord.ErrPetNotFound)
	}
	allowed := bypass || pet.OwnerPlayerID == params.ActorPlayerID
	if !allowed {
		allowed, err = service.has(ctx, params.ActorPlayerID, petpolicy.ManageAny)
	}
	if err != nil || !allowed {
		return petrecord.Pet{}, firstError(err, petrecord.ErrNoRights)
	}
	if !bypass {
		err = service.checkInventoryLimit(ctx, pet.OwnerPlayerID)
	}
	if err != nil {
		return petrecord.Pet{}, err
	}
	picked := petrecord.Pet{}
	found = false
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if cancelErr := service.store.CancelBreedingPet(txCtx, pet.ID, params.RoomID); cancelErr != nil {
			return cancelErr
		}
		var pickupErr error
		picked, found, pickupErr = service.store.Pickup(txCtx, pet.ID, params.RoomID, pet.OwnerPlayerID, pet.Version)
		if pickupErr != nil || !found || hook == nil {
			return pickupErr
		}
		return hook(txCtx, picked)
	})
	if err != nil || !found {
		return petrecord.Pet{}, firstError(err, petrecord.ErrConflict)
	}
	service.runtime.RemovePlaced(params.RoomID, pet.ID)
	if active, activeFound := service.rooms.Find(params.RoomID); activeFound {
		if unit, removed := active.RemoveEntity(petruntime.EntityKey(pet.ID)); removed {
			service.runtime.ProjectRemove(ctx, active, unit.UnitID)
		}
	}
	service.runtime.SendInventoryAdd(ctx, picked.OwnerPlayerID, picked)
	service.runtime.Publish(ctx, petpickedup.Name, petpickedup.Payload{PetID: picked.ID, RoomID: params.RoomID, ActorPlayerID: params.ActorPlayerID})
	return picked, nil
}

// Move directs one visible pet through the existing room world.
func (service *Service) Move(ctx context.Context, roomID int64, petID int64, actorID int64, point grid.Point) error {
	pet, found := service.runtime.Snapshot(roomID, petID)
	if !found {
		return petrecord.ErrPetNotFound
	}
	allowed := pet.OwnerPlayerID == actorID
	var err error
	if !allowed {
		allowed, err = service.has(ctx, actorID, petpolicy.MoveAny)
	}
	if err != nil || !allowed {
		return firstError(err, petrecord.ErrNoRights)
	}
	active, found := service.rooms.Find(roomID)
	if !found {
		return petrecord.ErrInvalidState
	}
	err = service.runtime.MovePet(active, petID, point)
	if err != nil {
		if errors.Is(err, petrecord.ErrInvalidState) {
			return err
		}
		return petrecord.ErrTileNotFree
	}
	return nil
}

// Select records one actor's current pet selection.
func (service *Service) Select(roomID int64, petID int64, actorID int64) error {
	if !service.runtime.Select(roomID, petID, actorID) {
		return petrecord.ErrPetNotFound
	}
	return nil
}

// IsExpected reports whether an error is safe protocol feedback.
func IsExpected(err error) bool {
	return errors.Is(err, petrecord.ErrPetNotFound) || errors.Is(err, petrecord.ErrNoRights) || errors.Is(err, petrecord.ErrInventoryLimit) || errors.Is(err, petrecord.ErrRoomLimit) || errors.Is(err, petrecord.ErrTileNotFree) || errors.Is(err, petrecord.ErrPetsDisabled) || errors.Is(err, petrecord.ErrInvalidState) || errors.Is(err, petrecord.ErrConflict)
}

// firstError chooses an infrastructure failure before a domain fallback.
func firstError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}
