package routes

import (
	"github.com/gofiber/fiber/v2"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomaudit "github.com/niflaot/pixels/internal/realm/room/control/audit"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
	roompromotion "github.com/niflaot/pixels/internal/realm/room/promotion"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	netconn "github.com/niflaot/pixels/networking/connection"
	voteroutes "github.com/niflaot/pixels/pkg/http/room/routes/votes"
	wiredroutes "github.com/niflaot/pixels/pkg/http/room/routes/wired"
	"go.uber.org/fx"
)

// Dependencies contains focused room administration dependencies.
type Dependencies struct {
	// In marks dependencies for Fx injection.
	fx.In

	// Audit reads room rights and moderation history.
	Audit roomaudit.Manager
	// Moderation reads current room sanctions.
	Moderation roommoderation.Reader
	// Votes manages room upvotes.
	Votes roomvotes.Manager
	// Bundles manages room bundle templates.
	Bundles roombundle.Manager
	// ConfigRooms updates optimistic room settings.
	ConfigRooms roomservice.ConfigManager
	// Promotions manages active room event advertisements.
	Promotions *roompromotion.Service
	// WiredConfig stores WIRED execution limits.
	WiredConfig roomwired.Config
	// WiredStore persists WIRED nodes and snapshots.
	WiredStore record.Store
	// WiredRewards persists normalized reward definitions.
	WiredRewards record.RewardStore
	// WiredRegistry exposes the immutable behavior manifest.
	WiredRegistry *registry.Registry
	// WiredCompiler validates administrative settings.
	WiredCompiler *configuration.Compiler
	// WiredEngine owns active compiled room generations.
	WiredEngine *wiredruntime.Engine
	// WiredGames owns game lifecycle and scoreboards.
	WiredGames *game.Coordinator
}

const (
	// roomPath stores the room admin base path.
	roomPath = "/api/admin/rooms"

	// navigatorPath stores the navigator admin base path.
	navigatorPath = "/api/admin/navigator"
)

// Register mounts protected room and navigator administration routes.
func Register(app *fiber.App, rooms roomservice.Manager, runtime *roomlive.Registry, connections *netconn.Registry, navigator navservice.Manager, players *playerlive.Registry, entry *roomentry.Service, dependencies Dependencies) {
	app.Get(roomPath, listHandler(rooms))
	app.Get(roomPath+"/bundle-templates", bundleTemplatesHandler(dependencies.Bundles))
	app.Post(roomPath+"/:id/bundle-template", markBundleTemplateHandler(dependencies.Bundles))
	app.Delete(roomPath+"/:id/bundle-template", unmarkBundleTemplateHandler(dependencies.Bundles))
	app.Get(roomPath+"/:id", detailHandler(rooms))
	app.Get(roomPath+"/:id/occupancy", occupancyHandler(rooms, runtime))
	if dependencies.Promotions != nil {
		app.Get(roomPath+"/:id/promotion", promotionHandler(dependencies.Promotions))
		app.Delete(roomPath+"/:id/promotion", cancelPromotionHandler(dependencies.Promotions))
	}
	if dependencies.ConfigRooms != nil {
		app.Patch(roomPath+"/:id/roller", rollerSettingsHandler(dependencies.ConfigRooms, runtime))
	}
	app.Post(roomPath+"/:id/close", closeHandler(runtime))
	app.Post(roomPath+"/:id/forward", forwardHandler(runtime, connections))
	app.Post(roomPath+"/players/:playerId/teleport", teleportHandler(rooms, players, connections, entry))
	app.Get(navigatorPath+"/categories", categoriesHandler(rooms))
	app.Get(navigatorPath+"/lifted", liftedHandler(navigator))
	app.Get(navigatorPath+"/history", navigatorHistoryHandler(navigator))
	app.Delete(navigatorPath+"/history", deleteNavigatorHistoryHandler(navigator))
	app.Get(navigatorPath+"/favorites", navigatorFavoritesHandler(navigator))
	app.Post(navigatorPath+"/official/:roomId", officialRoomHandler(dependencies.ConfigRooms, true))
	app.Delete(navigatorPath+"/official/:roomId", officialRoomHandler(dependencies.ConfigRooms, false))
	registerAudit(app, dependencies)
	voteroutes.Register(app, dependencies.Votes)
	wiredroutes.Register(app, wiredroutes.Dependencies{Config: dependencies.WiredConfig, Store: dependencies.WiredStore, Rewards: dependencies.WiredRewards, Registry: dependencies.WiredRegistry, Compiler: dependencies.WiredCompiler, Engine: dependencies.WiredEngine, Games: dependencies.WiredGames, Rooms: dependencies.ConfigRooms})
}
