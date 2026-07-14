package routes

import (
	"github.com/gofiber/fiber/v2"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/redis"
)

// Register mounts the authenticated administrative player routes.
func Register(app *fiber.App, players playerservice.AdminManager, redisClient *redis.Client, live *playerlive.Registry, connections *netconn.Registry, effectManagers ...playereffect.Manager) {
	handler := handler{players: players, idempotency: newIdempotencyStore(redisClient), live: live, connections: connections}
	if len(effectManagers) > 0 {
		handler.effects = effectManagers[0]
	}
	app.Post("/api/admin/players", handler.create)
	app.Get("/api/admin/players/by-username/:username", handler.readByUsername)
	app.Get("/api/admin/players/:id", handler.read)
	app.Patch("/api/admin/players/:id", handler.update)
	app.Delete("/api/admin/players/:id", handler.softDelete)
	app.Post("/api/admin/players/:playerId/effects", handler.grantEffect)
	app.Delete("/api/admin/players/:playerId/effects/:effectId", handler.revokeEffect)
}
