package commands

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	botsettings "github.com/niflaot/pixels/internal/realm/bot/settings"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outconfiguration "github.com/niflaot/pixels/networking/outbound/bot/configuration"
	outcontextmenu "github.com/niflaot/pixels/networking/outbound/bot/contextmenu"
	outskilllist "github.com/niflaot/pixels/networking/outbound/bot/skilllist"
)

const (
	// ConfigurationName identifies a bot configuration read.
	ConfigurationName command.Name = "bot.configuration"
	// SkillSaveName identifies a bot skill mutation.
	SkillSaveName command.Name = "bot.skill_save"
)

// ConfigurationCommand reads one bot skill value.
type ConfigurationCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// BotID identifies the configured bot.
	BotID int64
	// SkillID identifies the requested skill.
	SkillID int32
}

// CommandName returns the stable command name.
func (ConfigurationCommand) CommandName() command.Name { return ConfigurationName }

// SkillSaveCommand changes one bot skill.
type SkillSaveCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// BotID identifies the configured bot.
	BotID int64
	// SkillID identifies the changed skill.
	SkillID int32
	// Data stores the client configuration payload.
	Data string
}

// CommandName returns the stable command name.
func (SkillSaveCommand) CommandName() command.Name { return SkillSaveName }

// SettingsHandler handles bot configuration reads and writes.
type SettingsHandler struct {
	// Service coordinates bot behavior.
	Service *botsettings.Service
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connections.
	Bindings *binding.Registry
}

// Handle sends current configuration for one skill.
func (handler SettingsHandler) Handle(ctx context.Context, envelope command.Envelope[ConfigurationCommand]) error {
	resolved, roomID, err := handler.context(envelope.Command.Handler)
	if err != nil {
		return err
	}
	data, err := handler.Service.ConfigurationData(ctx, roomID, envelope.Command.BotID, resolved.ID(), envelope.Command.SkillID)
	if isExpected(err) {
		return PlacementHandler{}.softError(ctx, envelope.Command.Handler, err)
	}
	if err != nil {
		return err
	}
	packet, err := outconfiguration.Encode(envelope.Command.BotID, envelope.Command.SkillID, data)
	if err != nil {
		return err
	}
	return envelope.Command.Handler.Send(ctx, packet)
}

// HandleSave persists one bot skill and refreshes Nitro's context menu.
func (handler SettingsHandler) HandleSave(ctx context.Context, envelope command.Envelope[SkillSaveCommand]) error {
	resolved, roomID, err := handler.context(envelope.Command.Handler)
	if err != nil {
		return err
	}
	saved, err := handler.Service.SaveSkill(ctx, roomID, envelope.Command.BotID, resolved.ID(), envelope.Command.SkillID, envelope.Command.Data)
	if isExpected(err) {
		return PlacementHandler{}.softError(ctx, envelope.Command.Handler, err)
	}
	if err != nil {
		return err
	}
	packet, err := outskilllist.Encode(saved.ID, botcore.SkillRecords(saved))
	if err != nil {
		return err
	}
	if err = envelope.Command.Handler.Send(ctx, packet); err != nil {
		return err
	}
	packet, err = outcontextmenu.Encode(saved.ID)
	if err != nil {
		return err
	}
	return envelope.Command.Handler.Send(ctx, packet)
}

// context resolves the actor and current room.
func (handler SettingsHandler) context(connection netconn.Context) (*playerlive.Player, int64, error) {
	resolved, err := player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return nil, 0, err
	}
	roomID, found := resolved.CurrentRoom()
	if !found {
		return nil, 0, botrecord.ErrRoomNotFound
	}
	return resolved, roomID, nil
}
