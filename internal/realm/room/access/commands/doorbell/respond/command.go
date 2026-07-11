// Package respond resolves room doorbell entry requests.
package respond

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	entercmd "github.com/niflaot/pixels/internal/realm/room/access/commands/enter"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomsession "github.com/niflaot/pixels/internal/realm/room/runtime/commands/session"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outdoorbelldenied "github.com/niflaot/pixels/networking/outbound/room/doorbell/denied"
	outdoorbellhide "github.com/niflaot/pixels/networking/outbound/room/doorbell/hide"
)

const (
	// Name identifies the room doorbell response command.
	Name command.Name = "room.doorbell.respond"
)

var (
	// ErrNotInRoom reports a responder without active room presence.
	ErrNotInRoom = errors.New("doorbell responder is not in a room")
	// ErrRequestNotFound reports a missing waiting request.
	ErrRequestNotFound = errors.New("doorbell request not found")
)

// Command contains one room doorbell decision.
type Command struct {
	// Handler stores the responder connection context.
	Handler netconn.Context
	// Username identifies the waiting player.
	Username string
	// Accepted reports whether entry was approved.
	Accepted bool
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handler resolves room doorbell decisions.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Connections stores active connections.
	Connections *netconn.Registry
	// Entry decides room access and rights.
	Entry *roomentry.Service
	// Enter executes accepted room entry.
	Enter entercmd.Handler
}

// Handle resolves one room doorbell command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	responder, err := roomsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := responder.CurrentRoom()
	if !found {
		return ErrNotInRoom
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return ErrNotInRoom
	}
	snapshot := active.Snapshot()
	allowed, err := handler.Entry.CanAnswerDoorbell(ctx, roomID, snapshot.OwnerPlayerID, responder.ID())
	if err != nil {
		return err
	}
	if !allowed {
		return roomentry.ErrAccessDenied
	}
	waiting, found := active.ResolveDoorbell(envelope.Command.Username)
	if !found {
		return ErrRequestNotFound
	}
	if err := handler.notifyResponders(ctx, active, snapshot.OwnerPlayerID, waiting.Username, envelope.Command.Accepted); err != nil {
		return err
	}
	if !envelope.Command.Accepted {
		return sendDenied(ctx, waiting.Handler, "")
	}
	if err := sendAccepted(ctx, waiting.Handler, ""); err != nil {
		return err
	}

	return handler.Enter.Handle(ctx, command.Envelope[entercmd.Command]{Command: entercmd.Command{
		Handler: waiting.Handler, RoomID: roomID, Trusted: true,
	}})
}

// notifyResponders closes the doorbell prompt for every authorized responder.
func (handler Handler) notifyResponders(ctx context.Context, active *roomlive.Room, ownerPlayerID int64, username string, accepted bool) error {
	roomID := active.Snapshot().ID
	for _, occupant := range active.Occupants() {
		allowed, err := handler.Entry.CanAnswerDoorbell(ctx, roomID, ownerPlayerID, occupant.PlayerID)
		if err != nil {
			return err
		}
		if !allowed {
			continue
		}
		connection, found := handler.Connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		if !found {
			continue
		}
		if accepted {
			if err := sendAcceptedConnection(ctx, connection, username); err != nil {
				return err
			}
			continue
		}
		if err := sendDeniedConnection(ctx, connection, username); err != nil {
			return err
		}
	}

	return nil
}

// sendAccepted sends a doorbell acceptance through a handler context.
func sendAccepted(ctx context.Context, connection netconn.Context, username string) error {
	packet, err := outdoorbellhide.Encode(username)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// sendDenied sends a doorbell rejection through a handler context.
func sendDenied(ctx context.Context, connection netconn.Context, username string) error {
	packet, err := outdoorbelldenied.Encode(username)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// sendAcceptedConnection sends acceptance through a registered connection.
func sendAcceptedConnection(ctx context.Context, connection netconn.Connection, username string) error {
	packet, err := outdoorbellhide.Encode(username)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// sendDeniedConnection sends rejection through a registered connection.
func sendDeniedConnection(ctx context.Context, connection netconn.Connection, username string) error {
	packet, err := outdoorbelldenied.Encode(username)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
