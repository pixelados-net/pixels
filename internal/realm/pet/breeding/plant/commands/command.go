// Package commands owns monsterplant command workflows.
package commands

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	petplant "github.com/niflaot/pixels/internal/realm/pet/breeding/plant"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petsession "github.com/niflaot/pixels/internal/realm/pet/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// SupplementName identifies lifecycle supplements.
	SupplementName command.Name = "pet.plant.supplement"
	// HarvestName identifies mature plant harvesting.
	HarvestName command.Name = "pet.plant.harvest"
	// CompostName identifies dead plant composting.
	CompostName command.Name = "pet.plant.compost"
)

// Command stores one monsterplant request.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// Action identifies supplement, harvest, or compost.
	Action command.Name
	// PetID identifies the monsterplant.
	PetID int64
	// SupplementType identifies the lifecycle adjustment.
	SupplementType int32
}

// CommandName returns the selected stable command name.
func (value Command) CommandName() command.Name { return value.Action }

// Handler executes monsterplant requests.
type Handler struct {
	// Service owns lifecycle behavior.
	Service *petplant.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connections.
	Bindings *binding.Registry
}

// Handle executes one lifecycle request.
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
	case SupplementName:
		err = handler.Service.Supplement(ctx, roomID, envelope.Command.PetID, player.ID(), envelope.Command.SupplementType)
	case HarvestName:
		err = handler.Service.Harvest(ctx, envelope.Command.Handler, roomID, envelope.Command.PetID, player.ID())
	case CompostName:
		err = handler.Service.Compost(ctx, envelope.Command.Handler, roomID, envelope.Command.PetID, player.ID())
	}
	if errors.Is(err, petrecord.ErrNoRights) || errors.Is(err, petrecord.ErrInvalidState) || errors.Is(err, petrecord.ErrInvalidProduct) || errors.Is(err, petrecord.ErrConflict) {
		return nil
	}
	return err
}
