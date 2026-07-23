// Package commands owns pet presence command workflows.
package commands

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petpresence "github.com/niflaot/pixels/internal/realm/pet/presence"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petsession "github.com/niflaot/pixels/internal/realm/pet/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// InventoryName identifies pet inventory reads.
const InventoryName command.Name = "pet.inventory"

// InventoryCommand requests the authenticated owner's pet inventory.
type InventoryCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
}

// CommandName returns the stable command name.
func (InventoryCommand) CommandName() command.Name { return InventoryName }

// InventoryHandler handles pet inventory requests.
type InventoryHandler struct {
	// Service owns inventory reads.
	Service *petpresence.Service
	// Runtime owns fragmented protocol projection.
	Runtime InventoryProjector
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connections.
	Bindings *binding.Registry
}

// InventoryProjector sends one ordered inventory result.
type InventoryProjector interface {
	// SendInventory sends inventory fragments to one connection.
	SendInventory(context.Context, netconn.Context, []petrecord.Pet) error
}

// Handle sends the complete current inventory.
func (handler InventoryHandler) Handle(ctx context.Context, envelope command.Envelope[InventoryCommand]) error {
	player, err := petsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	pets, err := handler.Service.Inventory(ctx, player.ID())
	if err != nil {
		return err
	}
	return handler.Runtime.SendInventory(ctx, envelope.Command.Handler, pets)
}
