package wired

import (
	"github.com/gofiber/fiber/v2"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
)

// Dependencies contains WIRED administration boundaries.
type Dependencies struct {
	// Config stores validated feature limits.
	Config roomwired.Config
	// Store persists node configuration and snapshots.
	Store record.Store
	// Rewards persists normalized reward definitions.
	Rewards record.RewardStore
	// Registry exposes the immutable behavior manifest.
	Registry *registry.Registry
	// Compiler validates requests before persistence.
	Compiler *configuration.Compiler
	// Engine owns active room generations and traces.
	Engine *wiredruntime.Engine
	// Games owns administrative lifecycle transitions.
	Games *game.Coordinator
	// Rooms persists the room-level WIRED visibility setting.
	Rooms roomservice.ConfigManager
}

// Register mounts protected WIRED administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	if dependencies.Store == nil || dependencies.Registry == nil || dependencies.Compiler == nil || dependencies.Engine == nil {
		return
	}
	handler := Handler{dependencies: dependencies}
	app.Get("/api/admin/wired/registry", handler.registry)
	base := "/api/admin/rooms/:id/wired"
	app.Get(base, handler.room)
	app.Get(base+"/traces", handler.traces)
	app.Get(base+"/metrics", handler.metrics)
	app.Post(base+"/reload", handler.reload)
	app.Post(base+"/game/:action", handler.game)
	app.Put(base+"/visibility", handler.visibility)
	app.Get(base+"/:itemId", handler.item)
	app.Put(base+"/:itemId", handler.save)
	app.Post(base+"/:itemId/apply-snapshot", handler.snapshot)
	app.Get(base+"/:itemId/rewards", handler.rewards)
	app.Put(base+"/:itemId/rewards", handler.replaceRewards)
	app.Delete(base+"/:itemId/rewards", handler.deleteRewards)
}
