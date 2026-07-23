// Package handlers adapts bot packets to typed commands.
package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	botcommands "github.com/niflaot/pixels/internal/realm/bot/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbots "github.com/niflaot/pixels/networking/inbound/inventory/bots"
	"go.uber.org/zap"
)

// NewInventory creates the bot inventory packet handler.
func NewInventory(handler botcommands.InventoryHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		if err := inbots.Decode(packet); err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[botcommands.InventoryCommand]{Command: botcommands.InventoryCommand{Handler: connection}, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// RegisterInventory registers the bot inventory packet handler.
func RegisterInventory(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inbots.Header, handler)
}
