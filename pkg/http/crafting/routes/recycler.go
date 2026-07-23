package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	craftingpolicy "github.com/niflaot/pixels/internal/realm/crafting/policy"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

func (dependencies Dependencies) knownRecipes(ctx *fiber.Ctx) error {
	playerID, err := parseID(ctx, "playerId")
	if err != nil {
		return err
	}
	actor, err := readActor(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, actor, craftingpolicy.PlayerOverrideAny); err != nil {
		return err
	}
	known, err := dependencies.Store.KnownRecipes(ctx.Context(), playerID)
	if err != nil {
		return err
	}
	return ctx.JSON(known)
}
func (dependencies Dependencies) rememberRecipe(ctx *fiber.Ctx) error {
	return dependencies.mutateKnowledge(ctx, true)
}
func (dependencies Dependencies) forgetRecipe(ctx *fiber.Ctx) error {
	return dependencies.mutateKnowledge(ctx, false)
}
func (dependencies Dependencies) mutateKnowledge(ctx *fiber.Ctx, remember bool) error {
	playerID, err := parseID(ctx, "playerId")
	if err != nil {
		return err
	}
	recipeID, err := parseID(ctx, "recipeId")
	if err != nil {
		return err
	}
	request, err := audit(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, craftingpolicy.PlayerOverrideAny); err != nil {
		return err
	}
	action := "recipe.forget"
	var changed bool
	if remember {
		changed, err = dependencies.Store.RememberRecipe(ctx.Context(), playerID, recipeID)
		action = "recipe.remember"
	} else {
		changed, err = dependencies.Store.ForgetRecipe(ctx.Context(), playerID, recipeID)
	}
	if err != nil {
		return err
	}
	if err = dependencies.writeAudit(ctx, request, action, "player_recipe", recipeID); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"changed": changed})
}
func (dependencies Dependencies) recyclerConfig(ctx *fiber.Ctx) error {
	actor, err := readActor(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, actor, craftingpolicy.RecyclerManageAny); err != nil {
		return err
	}
	return ctx.JSON(configResponse(dependencies.Recycler.Config()))
}
func (dependencies Dependencies) updateRecyclerConfig(ctx *fiber.Ctx) error {
	var request RecyclerConfigRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := dependencies.authorize(ctx, request.ActorPlayerID, craftingpolicy.RecyclerManageAny); err != nil {
		return err
	}
	config := dependencies.Recycler.Config()
	config.RecyclerEnabled = request.Enabled
	config.RecyclerBatchSize = request.BatchSize
	config.RecyclerRarityChance = request.RarityChance
	if err := dependencies.Recycler.UpdateConfig(config); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	if err := dependencies.writeAudit(ctx, request.AuditRequest, "recycler.config", "recycler", 0); err != nil {
		return err
	}
	return ctx.JSON(configResponse(config))
}
func (dependencies Dependencies) listPrizes(ctx *fiber.Ctx) error {
	actor, err := readActor(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, actor, craftingpolicy.RecyclerManageAny); err != nil {
		return err
	}
	prizes, err := dependencies.Store.Prizes(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.JSON(prizes)
}
func (dependencies Dependencies) addPrize(ctx *fiber.Ctx) error {
	var request PrizeRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	if err := dependencies.authorize(ctx, request.ActorPlayerID, craftingpolicy.RecyclerManageAny); err != nil {
		return err
	}
	if request.Tier < 1 || request.Tier > 5 || request.RewardDefinitionID <= 0 {
		return fiber.ErrUnprocessableEntity
	}
	changed, err := dependencies.Store.AddPrize(ctx.Context(), request.Tier, request.RewardDefinitionID)
	if err != nil {
		return err
	}
	if err = dependencies.writeAudit(ctx, request.AuditRequest, "prize.add", "prize", request.RewardDefinitionID); err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"changed": changed})
}
func (dependencies Dependencies) deletePrize(ctx *fiber.Ctx) error {
	tier, err := parseTier(ctx)
	if err != nil {
		return err
	}
	definitionID, err := parseID(ctx, "definitionId")
	if err != nil {
		return err
	}
	request, err := audit(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, craftingpolicy.RecyclerManageAny); err != nil {
		return err
	}
	changed, err := dependencies.Store.DeletePrize(ctx.Context(), tier, definitionID)
	if err != nil {
		return err
	}
	if !changed {
		return fiber.ErrNotFound
	}
	if err = dependencies.writeAudit(ctx, request, "prize.delete", "prize", definitionID); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
func (dependencies Dependencies) writeAudit(ctx *fiber.Ctx, request AuditRequest, action string, kind string, id int64) error {
	if request.Reason == "" {
		request.Reason = "administrative mutation"
	}
	return dependencies.Store.InsertAudit(ctx.Context(), craftingrecord.Audit{ActorPlayerID: request.ActorPlayerID, Action: action, EntityKind: kind, EntityID: id, Reason: request.Reason})
}
func readActor(ctx *fiber.Ctx) (int64, error) {
	value, err := strconv.ParseInt(ctx.Get("X-Actor-Player-ID"), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid administrative actor header")
	}
	return value, nil
}
