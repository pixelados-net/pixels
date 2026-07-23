package admin

import (
	"context"

	petidentity "github.com/niflaot/pixels/internal/realm/pet/identity"
	petcreated "github.com/niflaot/pixels/internal/realm/pet/identity/events/created"
	petdeleted "github.com/niflaot/pixels/internal/realm/pet/identity/events/deleted"
	petupdated "github.com/niflaot/pixels/internal/realm/pet/identity/events/updated"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// CreateParams stores one idempotent protected pet grant.
type CreateParams struct {
	// OwnerPlayerID identifies the receiving owner.
	OwnerPlayerID int64
	// Name stores the proposed visible name.
	Name string
	// TypeID identifies species.
	TypeID int32
	// BreedID identifies breed.
	BreedID int32
	// PaletteID identifies palette.
	PaletteID int32
	// Color stores renderer color.
	Color string
	// OperationKey makes retries idempotent.
	OperationKey string
	// Audit stores required attribution.
	Audit Audit
}

// Create grants one pet and records attribution atomically.
func (service *Service) Create(ctx context.Context, params CreateParams) (petrecord.Pet, bool, error) {
	if err := params.Audit.Validate(); err != nil || params.OwnerPlayerID <= 0 || params.OperationKey == "" {
		return petrecord.Pet{}, false, firstError(err, petrecord.ErrInvalidState)
	}
	pet := petrecord.Pet{}
	created := false
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var grantErr error
		pet, created, grantErr = service.catalog.Grant(txCtx, params.OwnerPlayerID, params.TypeID, params.BreedID, params.PaletteID, params.Color, params.Name, "admin:"+params.OperationKey)
		if grantErr != nil || !created {
			return grantErr
		}
		return service.audit(txCtx, pet.ID, params.Audit, "created")
	})
	if err == nil && created {
		service.runtime.SendInventoryAdd(ctx, pet.OwnerPlayerID, pet)
		service.runtime.SendInventoryReceived(ctx, pet.OwnerPlayerID, pet)
		service.runtime.Publish(ctx, petcreated.Name, petcreated.Payload{PetID: pet.ID, OwnerPlayerID: pet.OwnerPlayerID, TypeID: pet.TypeID})
	}
	return pet, created, err
}

// Update replaces protected mutable fields atomically and optimistically.
func (service *Service) Update(ctx context.Context, petID int64, patch petrecord.AdminPatch, audit Audit) (petrecord.Pet, error) {
	if err := audit.Validate(); err != nil || petID <= 0 || patch.Version <= 0 {
		return petrecord.Pet{}, firstError(err, petrecord.ErrInvalidState)
	}
	if patch.Name != nil {
		name, code := service.catalog.ValidateName(*patch.Name)
		if code != petidentity.NameApproved {
			return petrecord.Pet{}, petrecord.ErrInvalidName
		}
		patch.Name = &name
	}
	if patch.Color != nil {
		color, err := petidentity.NormalizeColor(*patch.Color)
		if err != nil {
			return petrecord.Pet{}, err
		}
		patch.Color = &color
	}
	var saved petrecord.Pet
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var updated bool
		var updateErr error
		saved, updated, updateErr = service.store.UpdateAdmin(txCtx, petID, patch)
		if updateErr != nil || !updated {
			return firstError(updateErr, petrecord.ErrConflict)
		}
		return service.audit(txCtx, petID, audit, "updated")
	})
	if err == nil {
		service.projectMutation(ctx, saved)
		service.runtime.Publish(ctx, petupdated.Name, petupdated.Payload{PetID: saved.ID, Version: saved.Version})
	}
	return saved, err
}

// TransferOwner moves one inventory pet to a new owner.
func (service *Service) TransferOwner(ctx context.Context, petID int64, ownerID int64, version int64, audit Audit) (petrecord.Pet, error) {
	if err := audit.Validate(); err != nil || petID <= 0 || ownerID <= 0 || version <= 0 {
		return petrecord.Pet{}, firstError(err, petrecord.ErrInvalidState)
	}
	before, found, err := service.store.Find(ctx, petID)
	if err != nil || !found {
		return petrecord.Pet{}, firstError(err, petrecord.ErrPetNotFound)
	}
	var saved petrecord.Pet
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var updated bool
		saved, updated, err = service.store.TransferOwner(txCtx, petID, ownerID, version)
		if err != nil || !updated {
			return firstError(err, petrecord.ErrConflict)
		}
		return service.audit(txCtx, petID, audit, "owner_transferred")
	})
	if err == nil {
		service.runtime.SendInventoryRemove(ctx, before.OwnerPlayerID, petID)
		service.runtime.SendInventoryAdd(ctx, ownerID, saved)
		service.runtime.Publish(ctx, petupdated.Name, petupdated.Payload{PetID: saved.ID, Version: saved.Version})
	}
	return saved, err
}

// SetLocation places or picks one pet through room-owned world validation.
func (service *Service) SetLocation(ctx context.Context, petID int64, roomID *int64, point *grid.Point, audit Audit) (petrecord.Pet, error) {
	if err := audit.Validate(); err != nil || petID <= 0 || roomID != nil && point == nil {
		return petrecord.Pet{}, firstError(err, petrecord.ErrInvalidState)
	}
	current, found, err := service.store.Find(ctx, petID)
	if err != nil || !found {
		return petrecord.Pet{}, firstError(err, petrecord.ErrPetNotFound)
	}
	var saved petrecord.Pet
	hook := func(txCtx context.Context, _ petrecord.Pet) error {
		return service.audit(txCtx, petID, audit, "location_changed")
	}
	if roomID == nil {
		if current.RoomID == nil {
			return current, nil
		}
		saved, err = service.presence.PickupAdmin(ctx, petID, *current.RoomID, hook)
	} else {
		saved, err = service.presence.PlaceAdmin(ctx, petID, *roomID, *point, hook)
	}
	if err != nil {
		return petrecord.Pet{}, err
	}
	service.runtime.Publish(ctx, petupdated.Name, petupdated.Payload{PetID: saved.ID, Version: saved.Version})
	return saved, nil
}

// UpdateStats applies bounded deltas and projects an active mutation.
func (service *Service) UpdateStats(ctx context.Context, petID int64, energy int32, happiness int32, experience int32, version int64, audit Audit) (petrecord.Pet, error) {
	if err := audit.Validate(); err != nil || petID <= 0 || version <= 0 {
		return petrecord.Pet{}, firstError(err, petrecord.ErrInvalidState)
	}
	before, found, err := service.store.Find(ctx, petID)
	if err != nil || !found {
		return petrecord.Pet{}, firstError(err, petrecord.ErrPetNotFound)
	}
	var saved petrecord.Pet
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var updated bool
		saved, updated, err = service.store.UpdateStats(txCtx, petID, energy, happiness, experience, version)
		if err != nil || !updated {
			return firstError(err, petrecord.ErrConflict)
		}
		return service.audit(txCtx, petID, audit, "stats_updated")
	})
	if err == nil && saved.RoomID != nil {
		if active, roomFound := service.rooms.Find(*saved.RoomID); roomFound {
			service.runtime.ProjectStatChange(ctx, active, before, saved, experience, false)
		}
	}
	if err == nil {
		service.runtime.Publish(ctx, petupdated.Name, petupdated.Payload{PetID: saved.ID, Version: saved.Version})
	}
	return saved, err
}

// Delete soft-deletes one pet, removes its unit, and records attribution.
func (service *Service) Delete(ctx context.Context, petID int64, version int64, audit Audit) error {
	if err := audit.Validate(); err != nil || petID <= 0 || version <= 0 {
		return firstError(err, petrecord.ErrInvalidState)
	}
	pet, found, err := service.store.Find(ctx, petID)
	if err != nil || !found {
		return firstError(err, petrecord.ErrPetNotFound)
	}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		deleted, deleteErr := service.store.DeleteAdmin(txCtx, petID, version)
		if deleteErr != nil || !deleted {
			return firstError(deleteErr, petrecord.ErrConflict)
		}
		return service.audit(txCtx, petID, audit, "deleted")
	})
	if err != nil {
		return err
	}
	if pet.RoomID != nil {
		service.runtime.RemovePlaced(*pet.RoomID, pet.ID)
		if active, roomFound := service.rooms.Find(*pet.RoomID); roomFound {
			if unit, removed := active.RemoveEntity(petruntime.EntityKey(pet.ID)); removed {
				service.runtime.ProjectRemove(ctx, active, unit.UnitID)
			}
		}
	} else {
		service.runtime.SendInventoryRemove(ctx, pet.OwnerPlayerID, pet.ID)
	}
	service.runtime.Publish(ctx, petdeleted.Name, petdeleted.Payload{PetID: pet.ID, OwnerPlayerID: pet.OwnerPlayerID})
	return nil
}

// projectMutation refreshes room or inventory projections after an update.
func (service *Service) projectMutation(ctx context.Context, pet petrecord.Pet) {
	if pet.RoomID != nil {
		service.runtime.ReplacePlaced(pet)
		if active, found := service.rooms.Find(*pet.RoomID); found {
			service.runtime.ProjectFigure(ctx, active, pet)
		}
		return
	}
	service.runtime.SendInventoryAdd(ctx, pet.OwnerPlayerID, pet)
}

// firstError chooses an infrastructure error before its domain fallback.
func firstError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}
