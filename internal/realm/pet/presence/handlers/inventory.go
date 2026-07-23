// Package handlers adapts pet presence packets to commands.
package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petcommands "github.com/niflaot/pixels/internal/realm/pet/presence/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlist "github.com/niflaot/pixels/networking/inbound/inventory/pet/list"
	"go.uber.org/zap"
)

// NewInventory creates the pet inventory packet adapter.
func NewInventory(handler petcommands.InventoryHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		if err := inlist.Decode(packet); err != nil {
			return err
		}
		value := petcommands.InventoryCommand{Handler: connection}
		return dispatcher.Dispatch(context.Background(), command.Envelope[petcommands.InventoryCommand]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// RegisterInventory registers pet inventory reads.
func RegisterInventory(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inlist.Header, handler)
}
