package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	petcommands "github.com/niflaot/pixels/internal/realm/pet/presence/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inpickup "github.com/niflaot/pixels/networking/inbound/room/pet/pickup"
	inplace "github.com/niflaot/pixels/networking/inbound/room/pet/place"
	"go.uber.org/zap"
)

// NewPlace creates the pet placement packet adapter.
func NewPlace(handler petcommands.PlacementHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inplace.Decode(packet)
		if err != nil {
			return err
		}
		value := petcommands.PlaceCommand{Handler: connection, PetID: payload.PetID, X: payload.X, Y: payload.Y}
		return dispatcher.Dispatch(context.Background(), command.Envelope[petcommands.PlaceCommand]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// NewPickup creates the pet pickup packet adapter.
func NewPickup(handler petcommands.PlacementHandler, log *zap.Logger) netconn.Handler {
	adapter := command.HandlerFunc[petcommands.PickupCommand](handler.HandlePickup)
	dispatcher, _ := command.NewDispatcher(adapter)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		petID, err := inpickup.Decode(packet)
		if err != nil {
			return err
		}
		value := petcommands.PickupCommand{Handler: connection, PetID: petID}
		return dispatcher.Dispatch(context.Background(), command.Envelope[petcommands.PickupCommand]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// RegisterPlacement registers pet place and pickup adapters.
func RegisterPlacement(registry *netconn.HandlerRegistry, place netconn.Handler, pickup netconn.Handler) {
	_ = registry.Register(inplace.Header, place)
	_ = registry.Register(inpickup.Header, pickup)
}
