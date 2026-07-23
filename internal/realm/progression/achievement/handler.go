package achievement

import (
	"context"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlimits "github.com/niflaot/pixels/networking/inbound/progression/achievement/limits"
	inlist "github.com/niflaot/pixels/networking/inbound/user/achievement/list"
	achievementdata "github.com/niflaot/pixels/networking/outbound/progression/achievement/data"
	outlimits "github.com/niflaot/pixels/networking/outbound/progression/achievement/limits"
	outlist "github.com/niflaot/pixels/networking/outbound/user/achievement/list"
)

// Handler adapts Nitro achievement catalog requests.
type Handler struct {
	// Catalog owns the immutable achievement definitions.
	Catalog *progressionengine.Catalog
	// Store loads durable player progress.
	Store progressionrecord.Store
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
}

// Register installs achievement list and threshold handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(inlist.Header, handler.list)
	_ = registry.Register(inlimits.Header, handler.limits)
}

// list sends every visible definition with the player's durable progress.
func (handler Handler) list(connection netconn.Context, packet codec.Packet) error {
	if err := inlist.Decode(packet); err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Catalog == nil || handler.Store == nil {
		return nil
	}
	progress, err := handler.Store.PlayerAchievements(context.Background(), playerID)
	if err != nil {
		return err
	}
	byDefinition := make(map[int64]progressionrecord.PlayerAchievement, len(progress))
	for _, value := range progress {
		byDefinition[value.DefinitionID] = value
	}
	generation := handler.Catalog.Current()
	if generation == nil {
		return nil
	}
	if err := handler.sendLimits(connection, generation); err != nil {
		return err
	}
	values := make([]achievementdata.Achievement, 0, len(generation.Catalog.Achievements))
	for _, definition := range generation.Catalog.Achievements {
		if definition.Visible && definition.Enabled {
			values = append(values, progressionengine.AchievementData(definition, byDefinition[definition.ID]))
		}
	}
	response, err := outlist.Encode(values, "")
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// limits sends all visible badge threshold groups.
func (handler Handler) limits(connection netconn.Context, packet codec.Packet) error {
	if err := inlimits.Decode(packet); err != nil {
		return err
	}
	if handler.Catalog == nil {
		return nil
	}
	generation := handler.Catalog.Current()
	if generation == nil {
		return nil
	}
	return handler.sendLimits(connection, generation)
}

// sendLimits sends the badge thresholds Nitro needs for description placeholders.
func (handler Handler) sendLimits(connection netconn.Context, generation *progressionengine.Generation) error {
	groups := make([]outlimits.Group, 0, len(generation.Catalog.Achievements))
	for _, definition := range generation.Catalog.Achievements {
		if !definition.Visible || !definition.Enabled {
			continue
		}
		group := outlimits.Group{Prefix: definition.Name, Levels: make([]outlimits.Level, 0, len(definition.Levels))}
		for _, level := range definition.Levels {
			group.Levels = append(group.Levels, outlimits.Level{Suffix: level.Level, Limit: progressionengine.Clamp(level.ProgressNeeded)})
		}
		groups = append(groups, group)
	}
	response, err := outlimits.Encode(groups)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// playerID resolves one authenticated connection binding.
func (handler Handler) playerID(connection netconn.Context) (int64, bool) {
	if handler.Bindings == nil {
		return 0, false
	}
	current, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	return current.PlayerID, found
}
