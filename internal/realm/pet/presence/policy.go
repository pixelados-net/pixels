package presence

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// canPlace resolves room rights and explicit placement bypasses.
func (service *Service) canPlace(ctx context.Context, active *roomlive.Room, playerID int64) (bool, error) {
	if active.CanManageFurniture(playerID) {
		return true, nil
	}
	return service.has(ctx, playerID, petpolicy.PlaceAny)
}

// checkRoomLimits validates global and per-owner room capacities.
func (service *Service) checkRoomLimits(ctx context.Context, roomID int64, ownerID int64) error {
	if service.runtime.PlacedCount(roomID) < service.config.MaxPerRoom && service.runtime.OwnerPlacedCount(roomID, ownerID) < service.config.MaxPerOwnerRoom {
		return nil
	}
	unlimited, err := service.has(ctx, ownerID, petpolicy.RoomLimitBypass)
	if err != nil {
		return err
	}
	if !unlimited {
		return petrecord.ErrRoomLimit
	}
	return nil
}

// checkInventoryLimit validates one owner's bounded pet inventory.
func (service *Service) checkInventoryLimit(ctx context.Context, ownerID int64) error {
	count, err := service.store.CountInventory(ctx, ownerID)
	if err != nil || count < service.config.MaxInventory {
		return err
	}
	unlimited, err := service.has(ctx, ownerID, petpolicy.InventoryLimitBypass)
	if err != nil {
		return err
	}
	if !unlimited {
		return petrecord.ErrInventoryLimit
	}
	return nil
}

// has resolves one optional permission checker.
func (service *Service) has(ctx context.Context, playerID int64, node permission.Node) (bool, error) {
	if service.permissions == nil {
		return false, nil
	}
	return service.permissions.HasPermission(ctx, playerID, node)
}
