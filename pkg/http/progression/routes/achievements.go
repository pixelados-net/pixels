package routes

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	progressionpolicy "github.com/niflaot/pixels/internal/realm/progression/policy"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	progressionrequest "github.com/niflaot/pixels/pkg/http/progression/routes/request"
)

// AchievementCreateRequest aliases the achievement creation payload.
type AchievementCreateRequest = progressionrequest.AchievementCreate

// AchievementUpdateRequest aliases the achievement update payload.
type AchievementUpdateRequest = progressionrequest.AchievementUpdate

// AchievementLevelRequest aliases the achievement level payload.
type AchievementLevelRequest = progressionrequest.AchievementLevel

// registerAchievements mounts achievement definition administration.
func registerAchievements(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath+"/achievements", dependencies.listAchievements)
	app.Post(basePath+"/achievements", dependencies.createAchievement)
	app.Patch(basePath+"/achievements/:id", dependencies.updateAchievement)
	app.Delete(basePath+"/achievements/:id", dependencies.disableAchievement)
	app.Post(basePath+"/achievements/:id/levels", dependencies.createAchievementLevel)
	app.Patch(basePath+"/achievements/:id/levels/:level", dependencies.updateAchievementLevel)
	app.Delete(basePath+"/achievements/:id/levels/:level", dependencies.deleteAchievementLevel)
}

// listAchievements returns the active immutable catalog generation.
func (dependencies Dependencies) listAchievements(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.ManageDefinitions); err != nil {
		return err
	}
	generation := dependencies.Catalog.Current()
	if generation == nil {
		return ctx.JSON([]progressionrecord.AchievementDefinition{})
	}
	return ctx.JSON(generation.Catalog.Achievements)
}

// createAchievement inserts one definition without hot-reloading the catalog.
func (dependencies Dependencies) createAchievement(ctx *fiber.Ctx) error {
	var request AchievementCreateRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	visible := true
	if request.Visible != nil {
		visible = *request.Visible
	}
	value := progressionrecord.AchievementDefinition{Name: strings.TrimSpace(request.Name), Category: strings.TrimSpace(request.Category), Subcategory: strings.TrimSpace(request.Subcategory), TriggerKey: strings.TrimSpace(request.TriggerKey), Visible: visible, Enabled: true}
	if value.Name == "" || value.Category == "" || value.TriggerKey == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "name, category and triggerKey are required")
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageDefinitions, request.Audit, "achievement.create", entityID("achievement", value.Name), func(txCtx context.Context) error {
		var createErr error
		value, createErr = dependencies.Admin.CreateAchievement(txCtx, value)
		return createErr
	})
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(value)
}

// updateAchievement replaces mutable fields under optimistic locking.
func (dependencies Dependencies) updateAchievement(ctx *fiber.Ctx) error {
	id, err := parsePositiveID(ctx, "id")
	if err != nil {
		return err
	}
	var request AchievementUpdateRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	value, found := dependencies.achievement(id)
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "achievement definition not found in current catalog")
	}
	applyAchievementUpdate(&value, request)
	err = dependencies.mutate(ctx, progressionpolicy.ManageDefinitions, request.Audit, "achievement.update", entityID("achievement", id), func(txCtx context.Context) error {
		value, err = dependencies.Admin.UpdateAchievement(txCtx, value, request.Version)
		return err
	})
	if err != nil {
		return err
	}
	return ctx.JSON(value)
}

// disableAchievement soft-disables one definition.
func (dependencies Dependencies) disableAchievement(ctx *fiber.Ctx) error {
	id, err := parsePositiveID(ctx, "id")
	if err != nil {
		return err
	}
	var request AuditRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, progressionpolicy.ManageDefinitions, request, "achievement.disable", entityID("achievement", id), func(txCtx context.Context) error {
		changed, disableErr := dependencies.Admin.DisableAchievement(txCtx, id)
		if disableErr == nil && !changed {
			return progressionrecord.ErrNotFound
		}
		return disableErr
	})
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// createAchievementLevel adds one validated highest level.
func (dependencies Dependencies) createAchievementLevel(ctx *fiber.Ctx) error {
	return dependencies.writeAchievementLevel(ctx, false)
}

// updateAchievementLevel updates one existing cumulative level.
func (dependencies Dependencies) updateAchievementLevel(ctx *fiber.Ctx) error {
	return dependencies.writeAchievementLevel(ctx, true)
}

// writeAchievementLevel validates cumulative ordering and persists one level.
func (dependencies Dependencies) writeAchievementLevel(ctx *fiber.Ctx, update bool) error {
	id, err := parsePositiveID(ctx, "id")
	if err != nil {
		return err
	}
	var request AchievementLevelRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	level := request.Level
	if update {
		level, err = parsePositiveLevel(ctx)
	}
	value := progressionrecord.AchievementLevel{DefinitionID: id, Level: level, ProgressNeeded: request.ProgressNeeded, RewardCurrencyType: request.RewardCurrencyType, RewardAmount: request.RewardAmount, ScorePoints: request.ScorePoints}
	if err != nil || !dependencies.validLevel(value, update) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "achievement level or cumulative threshold is invalid")
	}
	action := "achievement.level.create"
	if update {
		action = "achievement.level.update"
	}
	err = dependencies.mutate(ctx, progressionpolicy.ManageDefinitions, request.Audit, action, entityID("achievement-level", id), func(txCtx context.Context) error { return dependencies.Admin.UpsertAchievementLevel(txCtx, value) })
	if err != nil {
		return err
	}
	return ctx.JSON(value)
}

// deleteAchievementLevel removes only one current highest level.
func (dependencies Dependencies) deleteAchievementLevel(ctx *fiber.Ctx) error {
	id, err := parsePositiveID(ctx, "id")
	if err != nil {
		return err
	}
	level, err := parsePositiveLevel(ctx)
	if err != nil {
		return err
	}
	var request AuditRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, progressionpolicy.ManageDefinitions, request, "achievement.level.delete", entityID("achievement-level", id), func(txCtx context.Context) error {
		_, deleteErr := dependencies.Admin.DeleteAchievementLevel(txCtx, id, level)
		return deleteErr
	})
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// achievement resolves one current cached definition.
func (dependencies Dependencies) achievement(id int64) (progressionrecord.AchievementDefinition, bool) {
	generation := dependencies.Catalog.Current()
	if generation == nil || generation.AchievementByID[id] == nil {
		return progressionrecord.AchievementDefinition{}, false
	}
	return *generation.AchievementByID[id], true
}

// applyAchievementUpdate applies optional request fields.
func applyAchievementUpdate(value *progressionrecord.AchievementDefinition, request AchievementUpdateRequest) {
	if request.Category != nil {
		value.Category = strings.TrimSpace(*request.Category)
	}
	if request.Subcategory != nil {
		value.Subcategory = strings.TrimSpace(*request.Subcategory)
	}
	if request.TriggerKey != nil {
		value.TriggerKey = strings.TrimSpace(*request.TriggerKey)
	}
	if request.Visible != nil {
		value.Visible = *request.Visible
	}
	if request.Enabled != nil {
		value.Enabled = *request.Enabled
	}
}

// validLevel reports whether one level preserves cumulative ordering.
func (dependencies Dependencies) validLevel(value progressionrecord.AchievementLevel, update bool) bool {
	definition, found := dependencies.achievement(value.DefinitionID)
	if !found || value.Level <= 0 || value.ProgressNeeded <= 0 || value.RewardAmount < 0 || value.ScorePoints < 0 {
		return false
	}
	if !update && value.Level != int32(len(definition.Levels)+1) {
		return false
	}
	for _, level := range definition.Levels {
		if level.Level == value.Level-1 && level.ProgressNeeded >= value.ProgressNeeded {
			return false
		}
		if level.Level == value.Level+1 && level.ProgressNeeded <= value.ProgressNeeded {
			return false
		}
		if update && level.Level == value.Level {
			return true
		}
	}
	return !update
}
