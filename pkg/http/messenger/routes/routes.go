package routes

import (
	"github.com/gofiber/fiber/v2"
	messengerservice "github.com/niflaot/pixels/internal/realm/messenger/core"
	"github.com/niflaot/pixels/internal/realm/messenger/runtime/delivery"
	"go.uber.org/fx"
)

const basePath = "/api/admin/players/:playerId"

// Dependencies contains messenger administration collaborators.
type Dependencies struct {
	fx.In

	// Messenger manages durable messenger behavior.
	Messenger *messengerservice.Service `optional:"true"`
	// Delivery projects administrative removals to online players.
	Delivery *delivery.Sender `optional:"true"`
}

// Register registers protected messenger administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	if dependencies.Messenger == nil {
		return
	}
	app.Get(basePath+"/friends", friendsHandler(dependencies))
	app.Get(basePath+"/friends/requests", requestsHandler(dependencies))
	app.Delete(basePath+"/friends/:friendId", removeHandler(dependencies))
	app.Post(basePath+"/privacy", privacyHandler(dependencies))
}
