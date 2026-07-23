package routes

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	progressionpolicy "github.com/niflaot/pixels/internal/realm/progression/policy"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	progressionrequest "github.com/niflaot/pixels/pkg/http/progression/routes/request"
)

// PlayerAchievementResponse joins player state with definition metadata.
type PlayerAchievementResponse struct {
	// DefinitionID identifies the achievement.
	DefinitionID int64 `json:"definitionId"`
	// Name stores the achievement group name.
	Name string `json:"name"`
	// Progress stores cumulative progress.
	Progress int64 `json:"progress"`
	// Level stores the highest paid level.
	Level int32 `json:"level"`
	// Badge stores the current achievement badge code.
	Badge string `json:"badge"`
}

// registerPlayers mounts player overrides and talent administration.
func registerPlayers(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath+"/players/:playerId/achievements", dependencies.playerAchievements)
	app.Post(basePath+"/players/:playerId/achievements/:definitionId/progress", dependencies.addPlayerProgress)
	app.Put(basePath+"/players/:playerId/achievements/:definitionId/level", dependencies.forcePlayerLevel)
	app.Delete(basePath+"/players/:playerId/achievements/:definitionId", dependencies.resetPlayerAchievement)
	app.Post(basePath+"/players/:playerId/badges", dependencies.grantPlayerBadge)
	app.Delete(basePath+"/players/:playerId/badges/:code", dependencies.removePlayerBadge)
	app.Put(basePath+"/talents/:track/levels/:level", dependencies.upsertTalentLevel)
	app.Get(basePath+"/players/:playerId/talents", dependencies.playerTalents)
	app.Put(basePath+"/players/:playerId/talents/:track", dependencies.forcePlayerTalent)
}

// playerAchievements lists player state joined to the current catalog.
func (dependencies Dependencies) playerAchievements(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.OverridePlayers); err != nil {
		return err
	}
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return err
	}
	progress, err := dependencies.Store.PlayerAchievements(ctx.Context(), playerID)
	if err != nil {
		return err
	}
	values := make([]PlayerAchievementResponse, 0, len(progress))
	for _, state := range progress {
		definition, found := dependencies.achievement(state.DefinitionID)
		if found {
			badge := ""
			if state.Level > 0 {
				badge = fmt.Sprintf("ACH_%s%d", definition.Name, state.Level)
			}
			values = append(values, PlayerAchievementResponse{DefinitionID: state.DefinitionID, Name: definition.Name, Progress: state.Progress, Level: state.Level, Badge: badge})
		}
	}
	return ctx.JSON(values)
}

// addPlayerProgress applies one positive delta through the gameplay engine.
func (dependencies Dependencies) addPlayerProgress(ctx *fiber.Ctx) error {
	playerID, definitionID, err := playerDefinitionIDs(ctx)
	if err != nil {
		return err
	}
	var request progressionrequest.Progress
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	if request.Amount <= 0 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "amount must be positive")
	}
	err = dependencies.mutate(ctx, progressionpolicy.OverridePlayers, request.Audit, "player.achievement.progress", entityID("player-achievement", playerID), func(txCtx context.Context) error {
		return dependencies.Engine.ProgressDefinition(txCtx, playerID, definitionID, request.Amount)
	})
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"progressed": true})
}

// forcePlayerLevel replaces one exact achievement level through the gameplay engine.
func (dependencies Dependencies) forcePlayerLevel(ctx *fiber.Ctx) error {
	playerID, definitionID, err := playerDefinitionIDs(ctx)
	if err != nil {
		return err
	}
	var request progressionrequest.ForceLevel
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, progressionpolicy.OverridePlayers, request.Audit, "player.achievement.level", entityID("player-achievement", playerID), func(txCtx context.Context) error {
		return dependencies.Engine.SetLevel(txCtx, playerID, definitionID, request.Level, request.PayRewards)
	})
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"level": request.Level})
}

// resetPlayerAchievement deletes one progress row without removing its badge.
func (dependencies Dependencies) resetPlayerAchievement(ctx *fiber.Ctx) error {
	playerID, definitionID, err := playerDefinitionIDs(ctx)
	if err != nil {
		return err
	}
	var request AuditRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, progressionpolicy.OverridePlayers, request, "player.achievement.reset", entityID("player-achievement", playerID), func(txCtx context.Context) error {
		_, resetErr := dependencies.Engine.ResetDefinition(txCtx, playerID, definitionID)
		return resetErr
	})
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// grantPlayerBadge grants one arbitrary badge code.
func (dependencies Dependencies) grantPlayerBadge(ctx *fiber.Ctx) error {
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return err
	}
	var request progressionrequest.Badge
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	code := strings.ToUpper(strings.TrimSpace(request.Badge))
	if code == "" || len(code) > 64 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid badge code")
	}
	granted := false
	err = dependencies.mutate(ctx, progressionpolicy.OverridePlayers, request.Audit, "player.badge.grant", entityID("player-badge", playerID), func(txCtx context.Context) error {
		var grantErr error
		granted, grantErr = dependencies.Badges.GrantBadge(txCtx, playerID, code, "admin")
		return grantErr
	})
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"badge": code, "granted": granted})
}

// removePlayerBadge removes one owned badge regardless of equipped state.
func (dependencies Dependencies) removePlayerBadge(ctx *fiber.Ctx) error {
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return err
	}
	code := strings.ToUpper(strings.TrimSpace(ctx.Params("code")))
	var request AuditRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, progressionpolicy.OverridePlayers, request, "player.badge.remove", entityID("player-badge", playerID), func(txCtx context.Context) error {
		removed, removeErr := dependencies.Badges.RemoveBadge(txCtx, playerID, code)
		if removeErr == nil && !removed {
			return progressionrecord.ErrNotFound
		}
		return removeErr
	})
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// upsertTalentLevel creates or replaces one track definition.
func (dependencies Dependencies) upsertTalentLevel(ctx *fiber.Ctx) error {
	level, err := parsePositiveLevel(ctx)
	if err != nil {
		return err
	}
	track := strings.TrimSpace(ctx.Params("track"))
	var request progressionrequest.TalentLevel
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	value := progressionrecord.TalentLevel{Track: track, Level: level, Requirements: request.Requirements, RewardItems: request.Items, RewardPerks: request.Perks, RewardBadges: request.Badges}
	if track == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "track is required")
	}
	err = dependencies.mutate(ctx, progressionpolicy.ManageDefinitions, request.Audit, "talent.level.upsert", entityID("talent", track), func(txCtx context.Context) error { return dependencies.Admin.UpsertTalentLevel(txCtx, value) })
	if err != nil {
		return err
	}
	return ctx.JSON(value)
}

// playerTalents lists one player's paid track levels.
func (dependencies Dependencies) playerTalents(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.OverridePlayers); err != nil {
		return err
	}
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return err
	}
	values, err := dependencies.Talents.PlayerLevels(ctx.Context(), playerID)
	if err != nil {
		return err
	}
	return ctx.JSON(values)
}

// forcePlayerTalent replaces one exact track level and rewards positive crossings.
func (dependencies Dependencies) forcePlayerTalent(ctx *fiber.Ctx) error {
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return err
	}
	track := strings.TrimSpace(ctx.Params("track"))
	var request progressionrequest.TalentForce
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, progressionpolicy.OverridePlayers, request.Audit, "player.talent.force", entityID("player-talent", playerID), func(txCtx context.Context) error {
		return dependencies.Talents.Force(txCtx, playerID, track, request.Level)
	})
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"track": track, "level": request.Level})
}

// playerDefinitionIDs parses one player and achievement path pair.
func playerDefinitionIDs(ctx *fiber.Ctx) (int64, int64, error) {
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return 0, 0, err
	}
	definitionID, err := parsePositiveID(ctx, "definitionId")
	return playerID, definitionID, err
}
