// Package routes exposes protected pet administration.
package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	petadmin "github.com/niflaot/pixels/internal/realm/pet/admin"
	"go.uber.org/fx"
)

const basePath = "/api/admin/pets"
const actorHeader = "X-Actor-Player-ID"

// Dependencies contains protected pet administration behavior.
type Dependencies struct {
	fx.In

	// Pets coordinates durable and live pet state.
	Pets *petadmin.Service
	// Permissions authorizes the attributed administrative actor.
	Permissions permissionservice.Checker
}

// Register mounts every protected pet administration route.
func Register(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath, dependencies.list)
	app.Post(basePath, dependencies.create)
	app.Get(basePath+"/metrics", dependencies.metrics)
	app.Get(basePath+"/species", dependencies.species)
	app.Get(basePath+"/breeds", dependencies.breeds)
	app.Get(basePath+"/commands", dependencies.commands)
	app.Post(basePath+"/reference/refresh", dependencies.refreshReference)
	app.Get(basePath+"/:id", dependencies.read)
	app.Patch(basePath+"/:id", dependencies.update)
	app.Delete(basePath+"/:id", dependencies.delete)
	app.Post(basePath+"/:id/owner", dependencies.transferOwner)
	app.Post(basePath+"/:id/location", dependencies.setLocation)
	app.Post(basePath+"/:id/stats", dependencies.updateStats)
}

// authorize verifies one positive actor against a registered pet node.
func (dependencies Dependencies) authorize(ctx *fiber.Ctx, actorPlayerID int64, node permission.Node) error {
	if actorPlayerID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid administrative actor")
	}
	allowed, err := dependencies.Permissions.HasPermission(ctx.Context(), actorPlayerID, node)
	if err != nil {
		return err
	}
	if !allowed {
		return fiber.NewError(fiber.StatusForbidden, "administrative actor lacks pet permission")
	}
	return nil
}

// readActor parses the required actor header used by read-only routes.
func readActor(ctx *fiber.Ctx) (int64, error) {
	actorPlayerID, err := strconv.ParseInt(ctx.Get(actorHeader), 10, 64)
	if err != nil || actorPlayerID <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid administrative actor header")
	}
	return actorPlayerID, nil
}

// authorizeRead verifies one read-only administrative request.
func (dependencies Dependencies) authorizeRead(ctx *fiber.Ctx, node permission.Node) error {
	actorPlayerID, err := readActor(ctx)
	if err != nil {
		return err
	}
	return dependencies.authorize(ctx, actorPlayerID, node)
}
