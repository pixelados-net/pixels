// Package routes contains protected WebSocket administration routes.
package routes

import (
	"github.com/gofiber/fiber/v2"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const basePath = "/api/admin/connections"

// Register mounts protected WebSocket administration routes.
func Register(app *fiber.App, registry *netconn.Registry) {
	app.Get(basePath, listHandler(registry))
	app.Get(basePath+"/list", listHandler(registry))
	app.Get(basePath+"/count", countHandler(registry))
	app.Get(basePath+"/reasons", reasonsHandler())
	app.Post(basePath+"/disconnect", disconnectAllHandler(registry))
	app.Post(basePath+"/:kind/disconnect", disconnectKindHandler(registry))
	app.Post(basePath+"/:kind/:id/disconnect", disconnectHandler(registry))
}
