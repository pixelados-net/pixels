package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	botcommands "github.com/niflaot/pixels/internal/realm/bot/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inpickup "github.com/niflaot/pixels/networking/inbound/bot/pickup"
	inplace "github.com/niflaot/pixels/networking/inbound/bot/place"
	"go.uber.org/zap"
)

// NewPlace creates the bot placement packet handler.
func NewPlace(handler botcommands.PlacementHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inplace.Decode(packet)
		if err != nil {
			return err
		}
		value := botcommands.PlaceCommand{Handler: connection, BotID: payload.BotID, X: payload.X, Y: payload.Y}
		return dispatcher.Dispatch(context.Background(), command.Envelope[botcommands.PlaceCommand]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// NewPickup creates the bot pickup packet handler.
func NewPickup(handler botcommands.PlacementHandler, log *zap.Logger) netconn.Handler {
	adapter := command.HandlerFunc[botcommands.PickupCommand](handler.HandlePickup)
	dispatcher, _ := command.NewDispatcher(adapter)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		botID, err := inpickup.Decode(packet)
		if err != nil {
			return err
		}
		value := botcommands.PickupCommand{Handler: connection, BotID: botID}
		return dispatcher.Dispatch(context.Background(), command.Envelope[botcommands.PickupCommand]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// RegisterPlacement registers bot place and pickup packet handlers.
func RegisterPlacement(registry *netconn.HandlerRegistry, place netconn.Handler, pickup netconn.Handler) {
	_ = registry.Register(inplace.Header, place)
	_ = registry.Register(inpickup.Header, pickup)
}
