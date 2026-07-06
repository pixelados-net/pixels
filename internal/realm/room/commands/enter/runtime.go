package enter

import (
	"context"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentered "github.com/niflaot/pixels/internal/realm/room/events/entered"
	roomleft "github.com/niflaot/pixels/internal/realm/room/events/left"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// join moves a player into the target room.
func (handler Handler) join(ctx context.Context, player *playerlive.Player, connection netconn.Context, room roommodel.Room) error {
	if previousID, found := player.CurrentRoom(); found && previousID != room.ID {
		if _, left, err := handler.Runtime.Leave(ctx, player.ID()); err != nil {
			return err
		} else if left {
			_ = handler.publish(ctx, roomleft.Name, roomleft.Payload{PlayerID: player.ID(), RoomID: previousID})
		}
	}

	_, err := handler.Runtime.Activate(roomSnapshot(room))
	if err != nil {
		return err
	}

	_, err = handler.Runtime.Join(ctx, room.ID, roomlive.Occupant{
		PlayerID:       player.ID(),
		Username:       player.Username(),
		ConnectionID:   connection.ConnectionID,
		ConnectionKind: connection.ConnectionKind,
	})
	if err != nil {
		return err
	}

	return handler.publish(ctx, roomentered.Name, roomentered.Payload{PlayerID: player.ID(), RoomID: room.ID})
}

// publish emits room lifecycle events.
func (handler Handler) publish(ctx context.Context, name bus.Name, payload any) error {
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{Name: name, Payload: payload})
}

// roomSnapshot maps persistent rooms to runtime snapshots.
func roomSnapshot(room roommodel.Room) roomlive.Snapshot {
	return roomlive.Snapshot{ID: room.ID, CategoryID: room.CategoryID, MaxUsers: room.MaxUsers}
}
