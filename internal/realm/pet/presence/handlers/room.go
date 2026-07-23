package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petcommands "github.com/niflaot/pixels/internal/realm/pet/presence/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininfo "github.com/niflaot/pixels/networking/inbound/room/pet/info"
	inmove "github.com/niflaot/pixels/networking/inbound/room/pet/move"
	inselect "github.com/niflaot/pixels/networking/inbound/room/pet/select"
	"go.uber.org/zap"
)

// NewInfo creates the pet information adapter.
func NewInfo(handler petcommands.RoomHandler, log *zap.Logger) netconn.Handler {
	return roomAdapter(handler, log, petcommands.InfoName, func(packet codec.Packet) (int64, int, int, error) {
		petID, err := ininfo.Decode(packet)
		return petID, 0, 0, err
	})
}

// NewMove creates the directed pet movement adapter.
func NewMove(handler petcommands.RoomHandler, log *zap.Logger) netconn.Handler {
	return roomAdapter(handler, log, petcommands.MoveName, func(packet codec.Packet) (int64, int, int, error) {
		payload, err := inmove.Decode(packet)
		return payload.PetID, payload.X, payload.Y, err
	})
}

// NewSelect creates the pet selection adapter.
func NewSelect(handler petcommands.RoomHandler, log *zap.Logger) netconn.Handler {
	return roomAdapter(handler, log, petcommands.SelectName, func(packet codec.Packet) (int64, int, int, error) {
		petID, err := inselect.Decode(packet)
		return petID, 0, 0, err
	})
}

// roomDecoder decodes one common room-pet request.
type roomDecoder func(codec.Packet) (int64, int, int, error)

// roomAdapter creates one common room-pet dispatcher.
func roomAdapter(handler petcommands.RoomHandler, log *zap.Logger, action command.Name, decode roomDecoder) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		petID, x, y, err := decode(packet)
		if err != nil {
			return err
		}
		value := petcommands.RoomCommand{Handler: connection, PetID: petID, Action: action, X: x, Y: y}
		return dispatcher.Dispatch(context.Background(), command.Envelope[petcommands.RoomCommand]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// RegisterRoom registers information, movement, and selection adapters.
func RegisterRoom(registry *netconn.HandlerRegistry, info netconn.Handler, move netconn.Handler, selected netconn.Handler) {
	_ = registry.Register(ininfo.Header, info)
	_ = registry.Register(inmove.Header, move)
	_ = registry.Register(inselect.Header, selected)
}
