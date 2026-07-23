package routes

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	craftingpolicy "github.com/niflaot/pixels/internal/realm/crafting/policy"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

func (dependencies Dependencies) listAltars(ctx *fiber.Ctx) error {
	actor, err := readActor(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, actor, craftingpolicy.AltarManageAny); err != nil {
		return err
	}
	altars, err := dependencies.Store.ListAltars(ctx.Context(), true)
	if err != nil {
		return err
	}
	response := make([]AltarResponse, 0, len(altars))
	for _, altar := range altars {
		recipes, readErr := dependencies.Store.Recipes(ctx.Context(), craftingrecord.RecipeFilter{AltarDefinitionID: altar.DefinitionID, IncludeUnknownSecrets: true, IncludeDisabled: true})
		if readErr != nil {
			return readErr
		}
		response = append(response, AltarResponse{Altar: altar, Recipes: recipes})
	}
	return ctx.JSON(response)
}
func (dependencies Dependencies) upsertAltar(ctx *fiber.Ctx) error {
	var request AltarRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := dependencies.authorize(ctx, request.ActorPlayerID, craftingpolicy.AltarManageAny); err != nil {
		return err
	}
	altar, created, err := dependencies.Store.UpsertAltar(ctx.Context(), request.DefinitionID)
	if err != nil {
		return err
	}
	if err = dependencies.writeAudit(ctx, request.AuditRequest, "altar.upsert", "altar", request.DefinitionID); err != nil {
		return err
	}
	status := fiber.StatusOK
	if created {
		status = fiber.StatusCreated
	}
	return ctx.Status(status).JSON(altar)
}
func (dependencies Dependencies) disableAltar(ctx *fiber.Ctx) error {
	id, err := parseID(ctx, "definitionId")
	if err != nil {
		return err
	}
	request, err := audit(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, craftingpolicy.AltarManageAny); err != nil {
		return err
	}
	changed, err := dependencies.Store.DisableAltar(ctx.Context(), id)
	if err != nil {
		return err
	}
	if !changed {
		return fiber.ErrNotFound
	}
	if err = dependencies.writeAudit(ctx, request, "altar.disable", "altar", id); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
func (dependencies Dependencies) createRecipe(ctx *fiber.Ctx) error {
	altarID, err := parseID(ctx, "definitionId")
	if err != nil {
		return err
	}
	var request RecipeCreateRequest
	if err = ctx.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, craftingpolicy.AltarManageAny); err != nil {
		return err
	}
	if strings.TrimSpace(request.Name) == "" || len(request.Ingredients) == 0 {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "recipe name and ingredients are required")
	}
	recipe, err := dependencies.Store.CreateRecipe(ctx.Context(), craftingrecord.CreateRecipe{AltarDefinitionID: altarID, Name: request.Name, RewardDefinitionID: request.RewardDefinitionID, Secret: request.Secret, Limited: request.Limited, Remaining: request.Remaining, AchievementCode: request.AchievementCode, Ingredients: request.Ingredients})
	if err != nil {
		return err
	}
	if err = dependencies.writeAudit(ctx, request.AuditRequest, "recipe.create", "recipe", recipe.ID); err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(recipe)
}
func (dependencies Dependencies) updateRecipe(ctx *fiber.Ctx) error {
	id, err := parseID(ctx, "recipeId")
	if err != nil {
		return err
	}
	var request RecipeUpdateRequest
	if err = ctx.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, craftingpolicy.AltarManageAny); err != nil {
		return err
	}
	recipe, changed, err := dependencies.Store.UpdateRecipe(ctx.Context(), id, request.RecipePatch)
	if errors.Is(err, craftingrecord.ErrConflict) {
		return fiber.ErrConflict
	}
	if err != nil {
		return err
	}
	if !changed {
		return fiber.ErrNotFound
	}
	if err = dependencies.writeAudit(ctx, request.AuditRequest, "recipe.update", "recipe", id); err != nil {
		return err
	}
	return ctx.JSON(recipe)
}
func (dependencies Dependencies) disableRecipe(ctx *fiber.Ctx) error {
	id, err := parseID(ctx, "recipeId")
	if err != nil {
		return err
	}
	request, err := audit(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, craftingpolicy.AltarManageAny); err != nil {
		return err
	}
	changed, err := dependencies.Store.DisableRecipe(ctx.Context(), id)
	if err != nil {
		return err
	}
	if !changed {
		return fiber.ErrNotFound
	}
	if err = dependencies.writeAudit(ctx, request, "recipe.disable", "recipe", id); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
