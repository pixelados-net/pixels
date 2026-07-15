package enter

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/leave"
	roomdoorbell "github.com/niflaot/pixels/internal/realm/room/access/doorbell"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	netconn "github.com/niflaot/pixels/networking/connection"
	outdoorbelladd "github.com/niflaot/pixels/networking/outbound/room/doorbell/add"
	outdoorbelldenied "github.com/niflaot/pixels/networking/outbound/room/doorbell/denied"
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
		if err := handler.loadRights(ctx, active, room.ID); err != nil {
			return nil, err
		}
		if err := handler.loadMutes(ctx, active, room.ID); err != nil {
			return nil, err
		}
		if err := handler.loadWorld(ctx, active, room, roomLayout); err != nil {
			return nil, err
		}
	}
	snapshot := player.Snapshot()

	occupant := roomlive.Occupant{
		PlayerID:       player.ID(),
		Username:       player.Username(),
		Motto:          snapshot.Motto,
		Figure:         snapshot.Look,
		Gender:         string(snapshot.Gender),
		ConnectionID:   connection.ConnectionID,
		ConnectionKind: connection.ConnectionKind,
	}
	if snapshot.ActiveEffectID != nil {
		occupant.ActiveEffectID = *snapshot.ActiveEffectID
	}
	_, err = handler.Runtime.Join(ctx, room.ID, occupant)
	if err == roomlive.ErrRoomFull && handler.Entry != nil {
		allowed, permissionErr := handler.Entry.CanEnterFull(ctx, player.ID())
		if permissionErr != nil {
			return nil, permissionErr
		}
		if allowed {
			_, err = handler.Runtime.JoinWithCapacity(ctx, room.ID, occupant, true)
		}
	}
	if err != nil {
		return nil, err
	}

	return active, nil
}

// loadMutes projects persistent active mutes into a newly loaded room.
func (handler Handler) loadMutes(ctx context.Context, room *roomlive.Room, roomID int64) error {
	if handler.Moderation == nil {
		room.ReplaceMutes(nil)
		return nil
	}
	mutes, err := handler.Moderation.ListMutes(ctx, roomID)
	if err != nil {
		return err
	}
	projected := make(map[int64]time.Time, len(mutes))
	for _, mute := range mutes {
		projected[mute.PlayerID] = mute.EndsAt
	}
	room.ReplaceMutes(projected)

	return nil
}

// loadRights projects persistent build rights into an active room.
func (handler Handler) loadRights(ctx context.Context, room *roomlive.Room, roomID int64) error {
	if handler.Rights == nil {
		room.ReplaceRights(nil)
		return nil
	}
	rights, err := handler.Rights.ListRights(ctx, roomID)
	if err != nil {
		return err
	}
	playerIDs := make([]int64, len(rights))
	for index := range rights {
		playerIDs[index] = rights[index].PlayerID
	}
	room.ReplaceRights(playerIDs)

	return nil
}

// requestDoorbell queues a player and notifies every authorized responder.
func (handler Handler) requestDoorbell(ctx context.Context, player *playerlive.Player, connection netconn.Context, room roommodel.Room) error {
	active, found := handler.Runtime.Find(room.ID)
	if !found {
		return handler.sendDoorbellDenied(ctx, connection)
	}
	approvers, err := handler.doorbellApprovers(ctx, active, room)
	if err != nil {
		return err
	}
	entry := roomdoorbell.Entry{
		PlayerID: player.ID(), Username: player.Username(), Handler: connection, RequestedAt: time.Now(),
	}
	if !active.RequestDoorbell(entry, len(approvers) > 0) {
		return handler.sendDoorbellDenied(ctx, connection)
	}
	waiting, err := outdoorbelladd.Encode("")
	if err != nil {
		return err
	}
	if err := connection.Send(ctx, waiting); err != nil {
		return err
	}
	notice, err := outdoorbelladd.Encode(player.Username())
	if err != nil {
		return err
	}
	if handler.Connections == nil {
		active.ResolveDoorbell(player.Username())

		return handler.sendDoorbellDenied(ctx, connection)
	}
	delivered := false
	for _, occupant := range approvers {
		approver, found := handler.Connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		if found && approver.Send(ctx, notice) == nil {
			delivered = true
		}
	}
	if delivered {
		return nil
	}
	active.ResolveDoorbell(player.Username())

	return handler.sendDoorbellDenied(ctx, connection)
}

// doorbellApprovers returns room occupants allowed to resolve waiting requests.
func (handler Handler) doorbellApprovers(ctx context.Context, active *roomlive.Room, room roommodel.Room) ([]roomlive.Occupant, error) {
	occupants := active.Occupants()
	approvers := make([]roomlive.Occupant, 0, len(occupants))
	for _, occupant := range occupants {
		allowed := occupant.PlayerID == room.OwnerPlayerID
		if !allowed && handler.Entry != nil {
			var err error
			allowed, err = handler.Entry.CanAnswerDoorbell(ctx, room.ID, room.OwnerPlayerID, occupant.PlayerID)
			if err != nil {
				return nil, err
			}
		}
		if allowed {
			approvers = append(approvers, occupant)
		}
	}

	return approvers, nil
}

// sendDoorbellDenied rejects a waiting room entry.
func (handler Handler) sendDoorbellDenied(ctx context.Context, connection netconn.Context) error {
	packet, err := outdoorbelldenied.Encode("")
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
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
