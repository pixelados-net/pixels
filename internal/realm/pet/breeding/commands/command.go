// Package commands owns pet breeding command workflows.
package commands

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	petbreeding "github.com/niflaot/pixels/internal/realm/pet/breeding"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petsession "github.com/niflaot/pixels/internal/realm/pet/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// StartName identifies session start or owner confirmation.
	StartName command.Name = "pet.breeding.start"
	// CancelName identifies session cancellation.
	CancelName command.Name = "pet.breeding.cancel"
	// ConfirmName identifies offspring confirmation.
	ConfirmName command.Name = "pet.breeding.confirm"
)

// Command stores one breeding request.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// Action identifies start, cancel, or confirm.
	Action command.Name
	// NestItemID identifies the breeding nest.
	NestItemID int64
	// Name stores the proposed offspring name.
	Name string
	// PetOneID identifies the first parent.
	PetOneID int64
	// PetTwoID identifies the second parent.
	PetTwoID int64
}

// CommandName returns the selected stable command name.
func (value Command) CommandName() command.Name { return value.Action }

// Handler executes pet breeding workflows.
type Handler struct {
	// Service owns breeding behavior.
	Service *petbreeding.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connections.
	Bindings *binding.Registry
}

// Handle executes one breeding request.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := petsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, inRoom := player.CurrentRoom()
	if !inRoom {
		return nil
	}
	switch envelope.Command.Action {
	case StartName:
		err = handler.Service.Start(ctx, envelope.Command.Handler, roomID, player.ID(), envelope.Command.PetOneID, envelope.Command.PetTwoID)
	case CancelName:
		err = handler.Service.Cancel(ctx, envelope.Command.NestItemID, roomID)
	case ConfirmName:
		err = handler.Service.Confirm(ctx, envelope.Command.Handler, roomID, player.ID(), envelope.Command.NestItemID, envelope.Command.Name, envelope.Command.PetOneID, envelope.Command.PetTwoID)
	}
	if errors.Is(err, petrecord.ErrInvalidState) || errors.Is(err, petrecord.ErrNoRights) || errors.Is(err, petrecord.ErrConflict) {
		return nil
	}
	return err
}
