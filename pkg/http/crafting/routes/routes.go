// Package routes exposes protected crafting administration.
package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	craftingrecycler "github.com/niflaot/pixels/internal/realm/crafting/recycler"
	"go.uber.org/fx"
)

const basePath = "/api/admin/crafting"

// Dependencies contains protected crafting administration behavior.
type Dependencies struct {
	fx.In
	Store       craftingrecord.Store
	Recycler    *craftingrecycler.Service
	Permissions permissionservice.Checker
}

// Register mounts every protected crafting administration route.
func Register(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath+"/altars", dependencies.listAltars)
	app.Post(basePath+"/altars", dependencies.upsertAltar)
	app.Delete(basePath+"/altars/:definitionId", dependencies.disableAltar)
	app.Post(basePath+"/altars/:definitionId/recipes", dependencies.createRecipe)
	app.Patch(basePath+"/recipes/:recipeId", dependencies.updateRecipe)
	app.Delete(basePath+"/recipes/:recipeId", dependencies.disableRecipe)
	app.Get(basePath+"/players/:playerId/recipes", dependencies.knownRecipes)
	app.Post(basePath+"/players/:playerId/recipes/:recipeId", dependencies.rememberRecipe)
	app.Delete(basePath+"/players/:playerId/recipes/:recipeId", dependencies.forgetRecipe)
	app.Get(basePath+"/recycler/config", dependencies.recyclerConfig)
	app.Put(basePath+"/recycler/config", dependencies.updateRecyclerConfig)
	app.Get(basePath+"/recycler/prizes", dependencies.listPrizes)
	app.Post(basePath+"/recycler/prizes", dependencies.addPrize)
	app.Delete(basePath+"/recycler/prizes/:tier/:definitionId", dependencies.deletePrize)
}

func (dependencies Dependencies) authorize(ctx *fiber.Ctx, actorID int64, node permission.Node) error {
	if actorID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid administrative actor")
	}
	allowed, err := dependencies.Permissions.HasPermission(ctx.Context(), actorID, node)
	if err != nil {
		return err
	}
	if !allowed {
		return fiber.NewError(fiber.StatusForbidden, "administrative actor lacks crafting permission")
	}
	return nil
}
func parseID(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return value, nil
}
func parseTier(ctx *fiber.Ctx) (int32, error) {
	value, err := strconv.ParseInt(ctx.Params("tier"), 10, 32)
	if err != nil || value < 1 || value > 5 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid tier")
	}
	return int32(value), nil
}
func audit(ctx *fiber.Ctx) (AuditRequest, error) {
	var request AuditRequest
	if err := ctx.BodyParser(&request); err != nil {
		return request, fiber.NewError(fiber.StatusBadRequest, "invalid crafting request")
	}
	if request.ActorPlayerID <= 0 || request.Reason == "" {
		return request, fiber.NewError(fiber.StatusBadRequest, "actorPlayerId and reason are required")
	}
	return request, nil
}
