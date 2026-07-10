// Package routes contains protected catalog administration routes.
package routes

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	catalogadmin "github.com/niflaot/pixels/internal/realm/catalog/admin"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	// basePath stores the catalog administration route prefix.
	basePath = "/api/admin/catalog"
)

// Dependencies contains catalog administration collaborators.
type Dependencies struct {
	fx.In

	// Catalog manages persistent catalog administration.
	Catalog catalogadmin.Manager
	// Connections stores active protocol connections.
	Connections *netconn.Registry
	// Log records catalog publication delivery failures.
	Log *zap.Logger
}

// Register registers protected catalog administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	if dependencies.Log == nil {
		dependencies.Log = zap.NewNop()
	}
	app.Get(basePath+"/pages", pagesHandler(dependencies))
	app.Post(basePath+"/pages", createPageHandler(dependencies))
	app.Patch(basePath+"/pages/:id", updatePageHandler(dependencies))
	app.Get(basePath+"/items", itemsHandler(dependencies))
	app.Post(basePath+"/items", createItemHandler(dependencies))
	app.Patch(basePath+"/items/:id", updateItemHandler(dependencies))
	app.Delete(basePath+"/items/:id", deleteItemHandler(dependencies))
	app.Post(basePath+"/refresh", refreshHandler(dependencies))
	app.Get(basePath+"/sanitize-list", sanitizeListHandler(dependencies))
}

// catalogError maps administration errors to stable HTTP failures.
func catalogError(err error) error {
	switch {
	case errors.Is(err, catalogadmin.ErrPageNotFound), errors.Is(err, catalogadmin.ErrItemNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, catalogadmin.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, catalogadmin.ErrInvalidPage), errors.Is(err, catalogadmin.ErrInvalidItem),
		errors.Is(err, catalogadmin.ErrDefinitionNotFound), errors.Is(err, catalogadmin.ErrLimitedBelowSales):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	default:
		return err
	}
}
