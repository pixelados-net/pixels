package routes

import (
	"github.com/gofiber/fiber/v2"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// roomPath stores the room admin base path.
	roomPath = "/api/admin/rooms"

	// navigatorPath stores the navigator admin base path.
	navigatorPath = "/api/admin/navigator"
)

// Register mounts protected room and navigator administration routes.
func Register(app *fiber.App, rooms roomservice.Manager, runtime *roomlive.Registry, connections *netconn.Registry, navigator navservice.Manager) {
	app.Get(roomPath, listHandler(rooms))
	app.Get(roomPath+"/:id", detailHandler(rooms))
	app.Get(roomPath+"/:id/occupancy", occupancyHandler(rooms, runtime))
	app.Post(roomPath+"/:id/close", closeHandler(runtime))
	app.Post(roomPath+"/:id/forward", forwardHandler(runtime, connections))
	app.Get(navigatorPath+"/categories", categoriesHandler(rooms))
	app.Get(navigatorPath+"/lifted", liftedHandler(navigator))
}
