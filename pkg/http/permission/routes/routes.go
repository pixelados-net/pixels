// Package routes contains protected permission administration routes.
package routes

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"go.uber.org/fx"
)

const (
	// basePath stores the permission administration route prefix.
	basePath = "/api/admin/permissions"
)

// Dependencies contains permission administration collaborators.
type Dependencies struct {
	fx.In

	// Permissions manages permission groups, memberships, and grants.
	Permissions permissionservice.Manager
}

// Register registers protected permission administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	if dependencies.Permissions == nil {
		return
	}
	app.Get(basePath+"/nodes", nodesHandler())
	app.Get(basePath+"/groups", groupsHandler(dependencies))
	app.Post(basePath+"/groups", createGroupHandler(dependencies))
	app.Patch(basePath+"/groups/:id", updateGroupHandler(dependencies))
	app.Post(basePath+"/groups/:id/nodes", grantGroupNodeHandler(dependencies))
	app.Delete(basePath+"/groups/:id/nodes/:node", revokeGroupNodeHandler(dependencies))
	app.Post(basePath+"/players/:playerId/groups/:groupId", addPlayerGroupHandler(dependencies))
	app.Delete(basePath+"/players/:playerId/groups/:groupId", removePlayerGroupHandler(dependencies))
	app.Post(basePath+"/players/:playerId/nodes", grantPlayerNodeHandler(dependencies))
	app.Delete(basePath+"/players/:playerId/nodes/:node", revokePlayerNodeHandler(dependencies))
	app.Get(basePath+"/players/:playerId/effective", effectiveNodesHandler(dependencies))
	app.Get(basePath+"/players/:playerId/check", checkHandler(dependencies))
}

// permissionError maps domain errors to stable HTTP failures.
func permissionError(err error) error {
	switch {
	case errors.Is(err, permissionservice.ErrGroupNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, permissionservice.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, permissionservice.ErrInvalidPlayerID),
		errors.Is(err, permissionservice.ErrInvalidGroupID),
		errors.Is(err, permissionservice.ErrInvalidGroup),
		errors.Is(err, permissionservice.ErrInvalidNode),
		errors.Is(err, permissionservice.ErrInheritanceCycle):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	default:
		return err
	}
}
