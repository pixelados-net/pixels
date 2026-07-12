// Package use adapts furniture use packets to generic interaction commands.
package use

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	interactcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/interact"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incolorwheel "github.com/niflaot/pixels/networking/inbound/furniture/colorwheel"
	indiceactivate "github.com/niflaot/pixels/networking/inbound/furniture/dice/activate"
	indicedeactivate "github.com/niflaot/pixels/networking/inbound/furniture/dice/deactivate"
	inuse "github.com/niflaot/pixels/networking/inbound/furniture/use"
	"go.uber.org/zap"
)

// New creates the generic furniture use packet adapter.
func New(handler interactcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inuse.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[interactcmd.Command]{
			Command: interactcmd.Command{Handler: connection, ItemID: int64(payload.ItemID), State: payload.State},
		})
	}
}

// NewDiceActivate creates the dedicated dice activation packet adapter.
func NewDiceActivate(handler interactcmd.Handler, log *zap.Logger) netconn.Handler {
	return newItemAdapter(handler, log, interactcmd.ActionDice, func(packet codec.Packet) (int32, error) {
		payload, err := indiceactivate.Decode(packet)
		return payload.ItemID, err
	})
}

// NewDiceClose creates the dedicated dice close packet adapter.
func NewDiceClose(handler interactcmd.Handler, log *zap.Logger) netconn.Handler {
	return newItemAdapter(handler, log, interactcmd.ActionDiceClose, func(packet codec.Packet) (int32, error) {
		payload, err := indicedeactivate.Decode(packet)
		return payload.ItemID, err
	})
}

// NewColorWheel creates the dedicated color-wheel activation packet adapter.
func NewColorWheel(handler interactcmd.Handler, log *zap.Logger) netconn.Handler {
	return newItemAdapter(handler, log, interactcmd.ActionColorWheel, func(packet codec.Packet) (int32, error) {
		payload, err := incolorwheel.Decode(packet)
		return payload.ItemID, err
	})
}

// newItemAdapter creates a dedicated one-item furniture packet adapter.
func newItemAdapter(handler interactcmd.Handler, log *zap.Logger, action interactcmd.Action, decode func(codec.Packet) (int32, error)) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		itemID, err := decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[interactcmd.Command]{
			Command: interactcmd.Command{Handler: connection, ItemID: int64(itemID), Action: action},
		})
	}
}

// Register adds the generic furniture use handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inuse.Header, handler)
}

// RegisterDedicated adds Nitro's dedicated dice and color-wheel handlers.
func RegisterDedicated(registry *netconn.HandlerRegistry, dice netconn.Handler, closeDice netconn.Handler, colorWheel netconn.Handler) {
	_ = registry.Register(indiceactivate.Header, dice)
	_ = registry.Register(indicedeactivate.Header, closeDice)
	_ = registry.Register(incolorwheel.Header, colorWheel)
}
