// Package commands owns grouped bot command workflows.
package commands

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botlifecycle "github.com/niflaot/pixels/internal/realm/bot/lifecycle"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbots "github.com/niflaot/pixels/networking/outbound/inventory/bots/list"
)

const (
	// InventoryName identifies the bot inventory command.
	InventoryName command.Name = "bot.inventory"
)

// InventoryCommand requests the authenticated player's bot inventory.
type InventoryCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
}

// CommandName returns the stable command name.
func (InventoryCommand) CommandName() command.Name { return InventoryName }

// InventoryHandler handles bot inventory commands.
type InventoryHandler struct {
	// Service coordinates bot behavior.
	Service *botlifecycle.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connections.
	Bindings *binding.Registry
}

// Handle sends the complete bot inventory response.
func (handler InventoryHandler) Handle(ctx context.Context, envelope command.Envelope[InventoryCommand]) error {
	resolved, err := player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	bots, err := handler.Service.Inventory(ctx, resolved.ID())
	if err != nil {
		return err
	}
	values := make([]outbots.Bot, len(bots))
	for index, bot := range bots {
		values[index] = botcore.InventoryRecord(bot)
	}
	packet, err := outbots.Encode(values)
	if err != nil {
		return err
	}
	return envelope.Command.Handler.Send(ctx, packet)
}
