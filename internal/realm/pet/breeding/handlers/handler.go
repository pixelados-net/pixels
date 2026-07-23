// Package handlers adapts pet breeding packets to commands.
package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petcommands "github.com/niflaot/pixels/internal/realm/pet/breeding/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incancel "github.com/niflaot/pixels/networking/inbound/room/pet/breeding/cancel"
	inconfirm "github.com/niflaot/pixels/networking/inbound/room/pet/breeding/confirm"
	instart "github.com/niflaot/pixels/networking/inbound/room/pet/breeding/start"
	"go.uber.org/zap"
)

// NewStart creates breeding start adapter.
func NewStart(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, petcommands.StartName, func(packet codec.Packet) (petcommands.Command, error) {
		payload, err := instart.Decode(packet)
		return petcommands.Command{PetOneID: payload.PetOneID, PetTwoID: payload.PetTwoID}, err
	})
}

// NewCancel creates breeding cancellation adapter.
func NewCancel(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, petcommands.CancelName, func(packet codec.Packet) (petcommands.Command, error) {
		nestID, err := incancel.Decode(packet)
		return petcommands.Command{NestItemID: nestID}, err
	})
}

// NewConfirm creates offspring confirmation adapter.
func NewConfirm(handler petcommands.Handler, log *zap.Logger) netconn.Handler {
	return adapter(handler, log, petcommands.ConfirmName, func(packet codec.Packet) (petcommands.Command, error) {
		payload, err := inconfirm.Decode(packet)
		return petcommands.Command{NestItemID: payload.NestItemID, Name: payload.Name, PetOneID: payload.PetOneID, PetTwoID: payload.PetTwoID}, err
	})
}

// decoder decodes one breeding request.
type decoder func(codec.Packet) (petcommands.Command, error)

// adapter creates one breeding command dispatcher.
func adapter(handler petcommands.Handler, log *zap.Logger, action command.Name, decode decoder) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		value, err := decode(packet)
		if err != nil {
			return err
		}
		value.Handler, value.Action = connection, action
		return dispatcher.Dispatch(context.Background(), command.Envelope[petcommands.Command]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// Register installs every breeding packet adapter.
func Register(registry *netconn.HandlerRegistry, start netconn.Handler, cancel netconn.Handler, confirm netconn.Handler) {
	_ = registry.Register(instart.Header, start)
	_ = registry.Register(incancel.Header, cancel)
	_ = registry.Register(inconfirm.Header, confirm)
}
