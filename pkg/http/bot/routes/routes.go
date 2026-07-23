package routes

import (
	"github.com/gofiber/fiber/v2"
	botadmin "github.com/niflaot/pixels/internal/realm/bot/admin"
	"go.uber.org/fx"
)

const basePath = "/api/admin/bots"

// Dependencies contains bot administration behavior.
type Dependencies struct {
	fx.In
	// Bots coordinates durable and live bot state.
	Bots *botadmin.Service
}

// Register mounts protected bot administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath+"/serve-items", dependencies.listServeItems)
	app.Post(basePath+"/serve-items", dependencies.createServeItem)
	app.Patch(basePath+"/serve-items/:id", dependencies.updateServeItem)
	app.Delete(basePath+"/serve-items/:id", dependencies.deleteServeItem)
	app.Get(basePath+"/:id", dependencies.readBot)
	app.Post(basePath+"/:id/force-pickup", dependencies.forcePickup)
}
