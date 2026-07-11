package routes

import (
	"github.com/gofiber/fiber/v2"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomaudit "github.com/niflaot/pixels/internal/realm/room/control/audit"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	voteroutes "github.com/niflaot/pixels/pkg/http/room/routes/votes"
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
	app.Get(roomPath+"/:id", detailHandler(rooms))
	app.Get(roomPath+"/:id/occupancy", occupancyHandler(rooms, runtime))
	app.Post(roomPath+"/:id/close", closeHandler(runtime))
	app.Post(roomPath+"/:id/forward", forwardHandler(runtime, connections))
	app.Post(roomPath+"/players/:playerId/teleport", teleportHandler(rooms, players, connections, entry))
	app.Get(navigatorPath+"/categories", categoriesHandler(rooms))
	app.Get(navigatorPath+"/lifted", liftedHandler(navigator))
	registerAudit(app, dependencies)
	voteroutes.Register(app, dependencies.Votes)
}
