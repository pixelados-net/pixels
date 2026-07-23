package rights

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	rightsrevoked "github.com/niflaot/pixels/internal/realm/room/control/events/rightsrevoked"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/pkg/bus"
)

// authorize verifies room ownership or a global staff capability.
func (service *Service) authorize(ctx context.Context, roomID int64, actorID int64, ownNode permission.Node, anyNode permission.Node) (roommodel.Room, error) {
	if roomID <= 0 || actorID <= 0 {
		return roommodel.Room{}, ErrInvalidIdentity
	}
	room, found, err := service.rooms.FindByID(ctx, roomID)
	if err != nil {
		return roommodel.Room{}, err
	}
	if !found {
		return roommodel.Room{}, ErrRoomNotFound
	}
	allowed, err := service.hasPermission(ctx, actorID, anyNode)
	if err != nil || allowed {
		return room, err
	}
	if room.OwnerPlayerID != actorID {
		return roommodel.Room{}, ErrAccessDenied
	}
	allowed, err = service.hasPermission(ctx, actorID, ownNode)
	if err != nil {
		return roommodel.Room{}, err
	}
	if !allowed {
		return roommodel.Room{}, ErrAccessDenied
	}

	return room, nil
}

// hasPermission resolves one optional global permission node.
func (service *Service) hasPermission(ctx context.Context, playerID int64, node permission.Node) (bool, error) {
	if service.permissions == nil || node == "" {
		return false, nil
	}

	return service.permissions.HasPermission(ctx, playerID, node)
}

// revoke removes one rights holder and publishes its committed action.
func (service *Service) revoke(ctx context.Context, roomID int64, actorID int64, playerID int64, action rightsrevoked.Action) error {
	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		removed, err := service.store.Revoke(txCtx, roomID, playerID)
		if err != nil || !removed {
			return err
		}

		return service.publish(txCtx, rightsrevoked.Name, rightsrevoked.Payload{RoomID: roomID, PlayerID: playerID, ActorID: actorID, Action: action})
	})
}

// publish emits one rights mutation event.
func (service *Service) publish(ctx context.Context, name bus.Name, payload any) error {
	if service.events == nil {
		return nil
	}

	return service.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
}
