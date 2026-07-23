// Package votes exposes room vote administration routes.
package votes

import (
	"github.com/gofiber/fiber/v2"

	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
)

const (
	// Path stores the room vote administration base path.
	Path = "/api/admin/room-votes"
)

// Register mounts room vote administration routes.
func Register(app *fiber.App, votes roomvotes.Manager) {
	app.Get(Path+"/status", statusHandler(votes))
	app.Get(Path+"/list", listHandler(votes))
	app.Post(Path+"/cast", castHandler(votes))
}
