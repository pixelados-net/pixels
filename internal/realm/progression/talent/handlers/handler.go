// Package handlers adapts Nitro talent requests to derived progression tracks.
package handlers

import (
	"context"
	"fmt"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	progressiontalent "github.com/niflaot/pixels/internal/realm/progression/talent"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlevel "github.com/niflaot/pixels/networking/inbound/progression/talent/level"
	intrack "github.com/niflaot/pixels/networking/inbound/progression/talent/track"
	talentdata "github.com/niflaot/pixels/networking/outbound/progression/talent/data"
	outlevel "github.com/niflaot/pixels/networking/outbound/progression/talent/level"
	outtrack "github.com/niflaot/pixels/networking/outbound/progression/talent/track"
)

// Handler owns talent track requests.
type Handler struct {
	// Service derives and rewards talent levels.
	Service *progressiontalent.Service
	// Catalog resolves achievement prerequisite metadata.
	Catalog *progressionengine.Catalog
	// Store loads durable player progression.
	Store progressionrecord.Store
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
}

// Register installs talent track and level handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(intrack.Header, handler.track)
	_ = registry.Register(inlevel.Header, handler.level)
}

// track sends one complete derived talent track.
func (handler Handler) track(connection netconn.Context, packet codec.Packet) error {
	name, err := intrack.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Service == nil {
		return nil
	}
	ctx := context.Background()
	paid, achievements, err := handler.progress(ctx, playerID, name)
	if err != nil {
		return err
	}
	levels := handler.mapLevels(name, paid, achievements)
	response, err := outtrack.Encode(name, levels)
	if err != nil {
		return err
	}
	return connection.Send(ctx, response)
}

// level sends the player's paid and maximum level for one track.
func (handler Handler) level(connection netconn.Context, packet codec.Packet) error {
	name, err := inlevel.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Service == nil {
		return nil
	}
	paid, _, err := handler.progress(context.Background(), playerID, name)
	if err != nil {
		return err
	}
	response, err := outlevel.Encode(name, paid, int32(len(handler.Service.Levels(name))))
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// progress loads paid talent and achievement progress in two grouped reads.
func (handler Handler) progress(ctx context.Context, playerID int64, name string) (int32, map[int64]progressionrecord.PlayerAchievement, error) {
	talents, err := handler.Store.PlayerTalents(ctx, playerID)
	if err != nil {
		return 0, nil, err
	}
	paid := int32(0)
	for _, talent := range talents {
		if talent.Track == name {
			paid = talent.Level
			break
		}
	}
	rows, err := handler.Store.PlayerAchievements(ctx, playerID)
	if err != nil {
		return 0, nil, err
	}
	achievements := make(map[int64]progressionrecord.PlayerAchievement, len(rows))
	for _, row := range rows {
		achievements[row.DefinitionID] = row
	}
	return paid, achievements, nil
}

// mapLevels creates Nitro's nested talent state without additional persistence reads.
func (handler Handler) mapLevels(name string, paid int32, achievements map[int64]progressionrecord.PlayerAchievement) []talentdata.Level {
	levels := handler.Service.Levels(name)
	values := make([]talentdata.Level, 0, len(levels))
	for _, level := range levels {
		state := int32(0)
		if paid >= level.Level {
			state = 2
		} else if paid+1 == level.Level {
			state = 1
		}
		value := talentdata.Level{ID: level.Level, State: state, Perks: level.RewardPerks}
		value.Tasks = handler.tasks(level, state, achievements)
		for _, definitionID := range level.RewardItems {
			value.Products = append(value.Products, talentdata.Product{Name: "i", Value: progressionengine.Clamp(definitionID)})
		}
		values = append(values, value)
	}
	return values
}

// tasks maps achievement prerequisites for one track level.
func (handler Handler) tasks(level progressionrecord.TalentLevel, trackState int32, achievements map[int64]progressionrecord.PlayerAchievement) []talentdata.Task {
	values := make([]talentdata.Task, 0, len(level.Requirements))
	generation := handler.Catalog.Current()
	for _, requirement := range level.Requirements {
		definition := generation.AchievementByID[requirement.DefinitionID]
		if definition == nil || requirement.RequiredLevel <= 0 || int(requirement.RequiredLevel) > len(definition.Levels) {
			continue
		}
		progress := achievements[requirement.DefinitionID]
		state := int32(0)
		if trackState != 0 {
			state = 1
			if progress.Level >= requirement.RequiredLevel {
				state = 2
			}
		}
		values = append(values, talentdata.Task{ID: progressionengine.Clamp(definition.ID), Index: requirement.RequiredLevel, BadgeCode: fmt.Sprintf("ACH_%s%d", definition.Name, requirement.RequiredLevel), State: state, Progress: progressionengine.Clamp(progress.Progress), RequiredProgress: progressionengine.Clamp(definition.Levels[requirement.RequiredLevel-1].ProgressNeeded)})
	}
	return values
}

// playerID resolves one authenticated connection binding.
func (handler Handler) playerID(connection netconn.Context) (int64, bool) {
	if handler.Bindings == nil {
		return 0, false
	}
	current, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	return current.PlayerID, found
}
