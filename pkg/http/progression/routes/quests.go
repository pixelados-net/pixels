package routes

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	progressionpolicy "github.com/niflaot/pixels/internal/realm/progression/policy"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	progressionrequest "github.com/niflaot/pixels/pkg/http/progression/routes/request"
)

// PlayerQuestResponse returns one player's active quest and history.
type PlayerQuestResponse struct {
	// Active stores the current quest state when present.
	Active *progressionrecord.PlayerQuestState `json:"active"`
	// History stores all durable quest progress.
	History []progressionrecord.PlayerQuestState `json:"history"`
}

// registerQuests mounts campaign, quest, and player quest administration.
func registerQuests(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath+"/campaigns", dependencies.listCampaigns)
	app.Post(basePath+"/campaigns", dependencies.createCampaign)
	app.Patch(basePath+"/campaigns/:code", dependencies.updateCampaign)
	app.Delete(basePath+"/campaigns/:code", dependencies.disableCampaign)
	app.Get(basePath+"/quests", dependencies.listQuests)
	app.Post(basePath+"/quests", dependencies.createQuest)
	app.Patch(basePath+"/quests/:id", dependencies.updateQuest)
	app.Delete(basePath+"/quests/:id", dependencies.disableQuest)
	app.Get(basePath+"/players/:playerId/quests", dependencies.playerQuests)
	app.Post(basePath+"/players/:playerId/quests/:questId/complete", dependencies.completePlayerQuest)
	app.Delete(basePath+"/players/:playerId/quests/active", dependencies.cancelPlayerQuest)
}

// listCampaigns returns the current immutable campaign catalog.
func (dependencies Dependencies) listCampaigns(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.ManageQuests); err != nil {
		return err
	}
	generation := dependencies.Catalog.Current()
	if generation == nil {
		return ctx.JSON([]progressionrecord.QuestCampaign{})
	}
	return ctx.JSON(generation.Catalog.Campaigns)
}

// createCampaign inserts one campaign without hot-reloading the catalog.
func (dependencies Dependencies) createCampaign(ctx *fiber.Ctx) error {
	var request progressionrequest.Campaign
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	value := request.Value(request.Code)
	if !progressionrequest.ValidCampaign(value) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid campaign")
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageQuests, request.Audit, "campaign.create", entityID("campaign", value.Code), func(txCtx context.Context) error { return dependencies.Admin.CreateCampaign(txCtx, value) })
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(value)
}

// updateCampaign replaces one campaign.
func (dependencies Dependencies) updateCampaign(ctx *fiber.Ctx) error {
	var request progressionrequest.Campaign
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	value := request.Value(ctx.Params("code"))
	if !progressionrequest.ValidCampaign(value) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid campaign")
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageQuests, request.Audit, "campaign.update", entityID("campaign", value.Code), func(txCtx context.Context) error {
		changed, updateErr := dependencies.Admin.UpdateCampaign(txCtx, value)
		if updateErr == nil && !changed {
			return progressionrecord.ErrNotFound
		}
		return updateErr
	})
	if err != nil {
		return err
	}
	return ctx.JSON(value)
}

// disableCampaign soft-disables one campaign.
func (dependencies Dependencies) disableCampaign(ctx *fiber.Ctx) error {
	code := strings.TrimSpace(ctx.Params("code"))
	var request AuditRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageQuests, request, "campaign.disable", entityID("campaign", code), func(txCtx context.Context) error {
		changed, disableErr := dependencies.Admin.DisableCampaign(txCtx, code)
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

// listQuests returns the current immutable quest catalog.
func (dependencies Dependencies) listQuests(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.ManageQuests); err != nil {
		return err
	}
	generation := dependencies.Catalog.Current()
	if generation == nil {
		return ctx.JSON([]progressionrecord.QuestDefinition{})
	}
	return ctx.JSON(generation.Catalog.Quests)
}

// createQuest inserts one quest without hot-reloading the catalog.
func (dependencies Dependencies) createQuest(ctx *fiber.Ctx) error {
	var request progressionrequest.Quest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	value := request.Value(0)
	if !progressionrequest.ValidQuest(value) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid quest")
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageQuests, request.Audit, "quest.create", entityID("quest", value.Name), func(txCtx context.Context) error {
		var createErr error
		value, createErr = dependencies.Admin.CreateQuest(txCtx, value)
		return createErr
	})
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(value)
}

// updateQuest replaces one quest under optimistic locking.
func (dependencies Dependencies) updateQuest(ctx *fiber.Ctx) error {
	id, err := parsePositiveID(ctx, "id")
	if err != nil {
		return err
	}
	var request progressionrequest.Quest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	value := request.Value(id)
	if request.Version <= 0 || !progressionrequest.ValidQuest(value) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid quest or version")
	}
	err = dependencies.mutate(ctx, progressionpolicy.ManageQuests, request.Audit, "quest.update", entityID("quest", id), func(txCtx context.Context) error {
		value, err = dependencies.Admin.UpdateQuest(txCtx, value, request.Version)
		return err
	})
	if err != nil {
		return err
	}
	return ctx.JSON(value)
}

// disableQuest soft-disables one quest.
func (dependencies Dependencies) disableQuest(ctx *fiber.Ctx) error {
	id, err := parsePositiveID(ctx, "id")
	if err != nil {
		return err
	}
	var request AuditRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, progressionpolicy.ManageQuests, request, "quest.disable", entityID("quest", id), func(txCtx context.Context) error {
		changed, disableErr := dependencies.Admin.DisableQuest(txCtx, id)
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

// playerQuests returns one player's active quest and durable history.
func (dependencies Dependencies) playerQuests(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.OverridePlayers); err != nil {
		return err
	}
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return err
	}
	history, err := dependencies.Store.QuestProgress(ctx.Context(), playerID)
	if err != nil {
		return err
	}
	active, found, err := dependencies.Store.ActiveQuest(ctx.Context(), playerID)
	if err != nil {
		return err
	}
	response := PlayerQuestResponse{History: history}
	if found {
		response.Active = &active
	}
	return ctx.JSON(response)
}

// completePlayerQuest forces completion through the real reward workflow.
func (dependencies Dependencies) completePlayerQuest(ctx *fiber.Ctx) error {
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return err
	}
	questID, err := parsePositiveID(ctx, "questId")
	if err != nil {
		return err
	}
	var request AuditRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, progressionpolicy.OverridePlayers, request, "player.quest.complete", entityID("player-quest", playerID), func(txCtx context.Context) error { return dependencies.Quests.ForceComplete(txCtx, playerID, questID) })
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"completed": true})
}

// cancelPlayerQuest clears one player's active quest.
func (dependencies Dependencies) cancelPlayerQuest(ctx *fiber.Ctx) error {
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return err
	}
	var request AuditRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	err = dependencies.mutate(ctx, progressionpolicy.OverridePlayers, request, "player.quest.cancel", entityID("player-quest", playerID), func(txCtx context.Context) error { return dependencies.Quests.Cancel(txCtx, playerID, false) })
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
