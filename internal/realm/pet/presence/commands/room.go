package commands

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petpresence "github.com/niflaot/pixels/internal/realm/pet/presence"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	petsession "github.com/niflaot/pixels/internal/realm/pet/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// InfoName identifies pet information reads.
	InfoName command.Name = "pet.info"
	// MoveName identifies directed pet movement.
	MoveName command.Name = "pet.move"
	// SelectName identifies pet selection.
	SelectName command.Name = "pet.select"
)

// RoomCommand stores one room-pet request.
type RoomCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// PetID identifies the pet.
	PetID int64
	// Action identifies info, move, or select.
	Action command.Name
	// X stores an optional destination coordinate.
	X int
	// Y stores an optional destination coordinate.
	Y int
}

// CommandName returns the selected stable command name.
func (value RoomCommand) CommandName() command.Name { return value.Action }

// RoomHandler handles pet information, movement, and selection.
type RoomHandler struct {
	// Service owns room presence workflows.
	Service *petpresence.Service
	// Runtime owns visible protocol projection.
	Runtime *petruntime.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connections.
	Bindings *binding.Registry
}

// Handle executes one room-pet request.
func (handler RoomHandler) Handle(ctx context.Context, envelope command.Envelope[RoomCommand]) error {
	player, err := petsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, inRoom := player.CurrentRoom()
	if !inRoom {
		return nil
	}
	switch envelope.Command.Action {
	case InfoName:
		pet, found := handler.Runtime.Snapshot(roomID, envelope.Command.PetID)
		if !found {
			return nil
		}
		return handler.Runtime.SendInformation(ctx, envelope.Command.Handler, pet)
	case MoveName:
		point, valid := grid.NewPoint(envelope.Command.X, envelope.Command.Y)
		if !valid {
			return nil
		}
		err = handler.Service.Move(ctx, roomID, envelope.Command.PetID, player.ID(), point)
	case SelectName:
		err = handler.Service.Select(roomID, envelope.Command.PetID, player.ID())
	}
	if petpresence.IsExpected(err) {
		return nil
	}
	return err
}
