package enter

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/commands/leave"
	roomentered "github.com/niflaot/pixels/internal/realm/room/events/entered"
	"github.com/niflaot/pixels/internal/realm/room/layout"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// join moves a player into the target room.
func (handler Handler) join(ctx context.Context, player *playerlive.Player, connection netconn.Context, room roommodel.Room, roomLayout layout.Layout) (*roomlive.Room, error) {
	if previousID, found := player.CurrentRoom(); found && previousID != room.ID {
		if err := handler.leavePreviousRoom(ctx, player.ID()); err != nil {
			return nil, err
		}
	}

	active, err := handler.Runtime.Activate(roomSnapshot(room))
	if err != nil {
		return nil, err
	}
	if !active.WorldLoaded() {
		if err := loadWorld(active, roomLayout); err != nil {
			return nil, err
		}
	}
	snapshot := player.Snapshot()

	_, err = handler.Runtime.Join(ctx, room.ID, roomlive.Occupant{
		PlayerID:       player.ID(),
		Username:       player.Username(),
		Motto:          snapshot.Motto,
		Figure:         snapshot.Look,
		Gender:         string(snapshot.Gender),
		ConnectionID:   connection.ConnectionID,
		ConnectionKind: connection.ConnectionKind,
	})
	if err != nil {
		return nil, err
	}

	return active, handler.publish(ctx, roomentered.Name, roomentered.Payload{PlayerID: player.ID(), RoomID: room.ID})
}

// leavePreviousRoom runs the standard room leave command.
func (handler Handler) leavePreviousRoom(ctx context.Context, playerID int64) error {
	return (leavecmd.Handler{
		Players: handler.Players, Bindings: handler.Bindings, Runtime: handler.Runtime,
		Connections: handler.Connections, Events: handler.Events,
	}).Handle(ctx, command.Envelope[leavecmd.Command]{
		Command: leavecmd.Command{PlayerID: playerID},
	})
}

// loadWorld loads the room runtime world from its persistent layout.
func loadWorld(room *roomlive.Room, roomLayout layout.Layout) error {
	roomGrid, err := roomLayout.Grid()
	if err != nil {
		return err
	}
	doorPoint, ok := grid.NewPoint(roomLayout.DoorX, roomLayout.DoorY)
	if !ok {
		return roomlive.ErrInvalidWorld
	}

	return room.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid,
		Door: worldpath.Position{
			Point: doorPoint,
			Z:     grid.Height(roomLayout.DoorZ),
		},
		Body:  rotationFromLayout(roomLayout),
		Head:  rotationFromLayout(roomLayout),
		Rules: worldpath.DefaultRules(),
	})
}

// rotationFromLayout converts layout direction to runtime rotation.
func rotationFromLayout(roomLayout layout.Layout) worldunit.Rotation {
	return worldunit.Rotation(roomLayout.DoorDirection % 8)
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
