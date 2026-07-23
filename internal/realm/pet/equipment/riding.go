package equipment

import (
	"context"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// Mount attaches or detaches one actor from a visible pet.
func (service *Service) Mount(ctx context.Context, roomID int64, petID int64, actorID int64, mount bool) error {
	_, err := service.runtime.Mount(ctx, roomID, petID, actorID, mount)
	return err
}

// RemoveSaddle unequips one owned pet saddle and dismounts its rider.
func (service *Service) RemoveSaddle(ctx context.Context, roomID int64, petID int64, actorID int64) error {
	pet, found := service.runtime.Snapshot(roomID, petID)
	if !found || pet.OwnerPlayerID != actorID || !pet.HasSaddle {
		return petrecord.ErrNoRights
	}
	if rider, riding := service.runtime.Rider(roomID, petID); riding {
		_, _ = service.runtime.Mount(ctx, roomID, petID, rider, false)
	}
	saved, updated, err := service.store.SetSaddle(ctx, pet.ID, actorID, false, pet.Version)
	if err != nil || !updated {
		return firstError(err, petrecord.ErrConflict)
	}
	service.runtime.ReplacePlaced(saved)
	if active, activeFound := service.rooms.Find(roomID); activeFound {
		service.runtime.ProjectFigure(ctx, active, saved)
	}
	return nil
}

// TogglePublicRide toggles public riding for one owned pet.
func (service *Service) TogglePublicRide(ctx context.Context, roomID int64, petID int64, actorID int64) error {
	return service.toggleFlags(ctx, roomID, petID, actorID, true)
}

// TogglePublicBreed toggles public breeding for one owned pet.
func (service *Service) TogglePublicBreed(ctx context.Context, roomID int64, petID int64, actorID int64) error {
	return service.toggleFlags(ctx, roomID, petID, actorID, false)
}

// toggleFlags updates one public access flag with optimistic versioning.
func (service *Service) toggleFlags(ctx context.Context, roomID int64, petID int64, actorID int64, riding bool) error {
	pet, found := service.runtime.Snapshot(roomID, petID)
	if !found || pet.OwnerPlayerID != actorID {
		return petrecord.ErrNoRights
	}
	publicRide, publicBreed := pet.PublicRide, pet.PublicBreed
	if riding {
		publicRide = !publicRide
	} else {
		publicBreed = !publicBreed
	}
	saved, updated, err := service.store.UpdateFlags(ctx, pet.ID, actorID, publicRide, publicBreed, pet.Version)
	if err != nil || !updated {
		return firstError(err, petrecord.ErrConflict)
	}
	service.runtime.ReplacePlaced(saved)
	if active, activeFound := service.rooms.Find(roomID); activeFound {
		service.runtime.ProjectFigure(ctx, active, saved)
	}
	return nil
}
