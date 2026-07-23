// Package commands owns pet care command workflows.
package commands

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	petcare "github.com/niflaot/pixels/internal/realm/pet/care"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petsession "github.com/niflaot/pixels/internal/realm/pet/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outfailure "github.com/niflaot/pixels/networking/outbound/room/pet/respect/failure"
)

const (
	// RespectName identifies pet respect.
	RespectName command.Name = "pet.respect"
	// TrainingName identifies pet training-panel reads.
	TrainingName command.Name = "pet.training"
)

// Command stores one pet care request.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// PetID identifies the visible pet.
	PetID int64
	// Action identifies respect or training.
	Action command.Name
}

// CommandName returns the selected stable command name.
func (value Command) CommandName() command.Name { return value.Action }

// Handler executes pet care requests.
type Handler struct {
	// Service owns pet care behavior.
	Service *petcare.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connections.
	Bindings *binding.Registry
}

// Handle executes one care request without disconnecting on policy failures.
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
	case TrainingName:
		err = handler.Service.Training(ctx, envelope.Command.Handler, roomID, envelope.Command.PetID)
	case RespectName:
		var result petcare.RespectResult
		result, err = handler.Service.Respect(ctx, roomID, envelope.Command.PetID, player.ID())
		if result.TooYoung {
			packet, encodeErr := outfailure.Encode(result.AgeDays, result.RequiredAgeDays)
			if encodeErr != nil {
				return encodeErr
			}
			return envelope.Command.Handler.Send(ctx, packet)
		}
	}
	if errors.Is(err, petrecord.ErrPetNotFound) || errors.Is(err, petrecord.ErrNoRights) || errors.Is(err, petrecord.ErrRespectQuota) {
		return nil
	}
	return err
}
