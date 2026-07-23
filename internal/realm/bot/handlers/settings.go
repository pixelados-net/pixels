package handlers

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	botcommands "github.com/niflaot/pixels/internal/realm/bot/commands"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inconfiguration "github.com/niflaot/pixels/networking/inbound/bot/configuration"
	inskillsave "github.com/niflaot/pixels/networking/inbound/bot/skillsave"
	"go.uber.org/zap"
)

// NewConfiguration creates the bot configuration read handler.
func NewConfiguration(handler botcommands.SettingsHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inconfiguration.Decode(packet)
		if err != nil {
			return err
		}
		value := botcommands.ConfigurationCommand{Handler: connection, BotID: payload.BotID, SkillID: payload.SkillID}
		return dispatcher.Dispatch(context.Background(), command.Envelope[botcommands.ConfigurationCommand]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// NewSkillSave creates the bot skill mutation handler.
func NewSkillSave(handler botcommands.SettingsHandler, log *zap.Logger) netconn.Handler {
	adapter := command.HandlerFunc[botcommands.SkillSaveCommand](handler.HandleSave)
	dispatcher, _ := command.NewDispatcher(adapter)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inskillsave.Decode(packet)
		if err != nil {
			return err
		}
		value := botcommands.SkillSaveCommand{Handler: connection, BotID: payload.BotID, SkillID: payload.SkillID, Data: payload.Data}
		return dispatcher.Dispatch(context.Background(), command.Envelope[botcommands.SkillSaveCommand]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// RegisterSettings registers bot configuration packet handlers.
func RegisterSettings(registry *netconn.HandlerRegistry, configuration netconn.Handler, save netconn.Handler) {
	_ = registry.Register(inconfiguration.Header, configuration)
	_ = registry.Register(inskillsave.Header, save)
}
