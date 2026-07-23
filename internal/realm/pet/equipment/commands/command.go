// Package commands owns pet equipment command workflows.
package commands

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petequipment "github.com/niflaot/pixels/internal/realm/pet/equipment"
	petsession "github.com/niflaot/pixels/internal/realm/pet/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// ProductName identifies typed pet product use.
	ProductName command.Name = "pet.product.use"
	// HandItemName identifies virtual hand-item feeding.
	HandItemName command.Name = "pet.handitem.give"
	// MountName identifies mount or dismount.
	MountName command.Name = "pet.mount"
	// SaddleName identifies saddle removal.
	SaddleName command.Name = "pet.saddle.remove"
	// RidingName identifies public-riding toggles.
	RidingName command.Name = "pet.riding.toggle"
	// BreedingName identifies public-breeding toggles.
	BreedingName command.Name = "pet.breeding.toggle"
)

// Command stores one equipment request.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// Action identifies the equipment behavior.
	Action command.Name
	// PetID identifies a durable pet or room-local unit for hand items.
	PetID int64
	// ItemID identifies a placed furniture product.
	ItemID int64
	// Flag stores mount rather than dismount.
	Flag bool
}

// CommandName returns the selected stable command name.
func (value Command) CommandName() command.Name { return value.Action }

// Handler executes pet equipment requests.
type Handler struct {
	// Service owns pet equipment behavior.
	Service *petequipment.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connections.
	Bindings *binding.Registry
}

// Handle executes one equipment request.
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
	case ProductName:
		var result petequipment.ProductResult
		result, err = handler.Service.UseProduct(ctx, roomID, player.ID(), envelope.Command.ItemID, envelope.Command.PetID)
		if err == nil {
			err = handler.Service.SendProductInventoryChange(ctx, envelope.Command.Handler, envelope.Command.ItemID, result.Consumed)
		}
	case HandItemName:
		err = handler.Service.GiveHandItem(ctx, roomID, player.ID(), envelope.Command.PetID)
	case MountName:
		err = handler.Service.Mount(ctx, roomID, envelope.Command.PetID, player.ID(), envelope.Command.Flag)
	case SaddleName:
		err = handler.Service.RemoveSaddle(ctx, roomID, envelope.Command.PetID, player.ID())
	case RidingName:
		err = handler.Service.TogglePublicRide(ctx, roomID, envelope.Command.PetID, player.ID())
	case BreedingName:
		err = handler.Service.TogglePublicBreed(ctx, roomID, envelope.Command.PetID, player.ID())
	}
	if petequipment.IsExpected(err) {
		return nil
	}
	return err
}
