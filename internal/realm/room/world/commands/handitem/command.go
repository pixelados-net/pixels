// Package handitem coordinates carried room item drop and transfer.
package handitem

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outhand "github.com/niflaot/pixels/networking/outbound/room/entities/handitem"
	outreceived "github.com/niflaot/pixels/networking/outbound/room/entities/handitemreceived"
)

// Name identifies hand-item commands.
const Name command.Name = "room.handitem"

// Kind identifies one carried-item operation.
type Kind uint8

const (
	// KindDrop clears the actor's hand item.
	KindDrop Kind = iota + 1
	// KindGive transfers the actor's hand item.
	KindGive
)

// Command stores a carried-item request.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// Kind identifies drop or give.
	Kind Kind
	// TargetUnitID identifies a room-local recipient.
	TargetUnitID int64
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handler handles carried-item requests.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated session bindings.
	Bindings *binding.Registry
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Connections stores active transports.
	Connections *netconn.Registry
}

// Handle handles one carried-item request.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	bindingRecord, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: envelope.Command.Handler.ConnectionKind, ID: envelope.Command.Handler.ConnectionID})
	if !found {
		return nil
	}
	playerID := bindingRecord.PlayerID
	player, found := handler.Players.Find(playerID)
	if !found {
		return nil
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return nil
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return nil
	}
	actor, found := active.Unit(playerID)
	if !found || actor.HandItem <= 0 {
		return nil
	}
	if envelope.Command.Kind == KindDrop {
		return handler.setAndBroadcast(ctx, active, playerID, 0)
	}
	target, found := active.UnitByID(envelope.Command.TargetUnitID)
	if !found || target.PlayerID <= 0 || target.PlayerID == playerID || !adjacent(actor, target) {
		return nil
	}
	itemID := actor.HandItem
	if err := handler.setAndBroadcast(ctx, active, playerID, 0); err != nil {
		return err
	}
	if err := handler.notifyRecipient(ctx, active, target.PlayerID, actor.UnitID, itemID); err != nil {
		return err
	}
	return handler.setAndBroadcast(ctx, active, target.PlayerID, itemID)
}

// setAndBroadcast replaces one hand item and projects it room-wide.
func (handler Handler) setAndBroadcast(ctx context.Context, active *roomlive.Room, playerID int64, itemID int32) error {
	unit, err := active.SetHandItem(playerID, itemID)
	if err != nil {
		return err
	}
	packet, err := outhand.Encode(unit.UnitID, itemID)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
}

// notifyRecipient sends the private receipt chat event.
func (handler Handler) notifyRecipient(ctx context.Context, active *roomlive.Room, playerID int64, giverUnitID int64, itemID int32) error {
	occupant, found := active.Occupant(playerID)
	if !found {
		return nil
	}
	connection, found := handler.Connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
	if !found {
		return nil
	}
	packet, err := outreceived.Encode(giverUnitID, itemID)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// adjacent reports whether units share an edge or corner.
func adjacent(first roomlive.UnitSnapshot, second roomlive.UnitSnapshot) bool {
	dx := int(first.Position.Point.X) - int(second.Position.Point.X)
	if dx < 0 {
		dx = -dx
	}
	dy := int(first.Position.Point.Y) - int(second.Position.Point.Y)
	if dy < 0 {
		dy = -dy
	}
	return (dx != 0 || dy != 0) && dx <= 1 && dy <= 1
}
