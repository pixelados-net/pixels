// Package routes exposes protected chat administration endpoints.
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/realm/chat/bubble"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	"github.com/niflaot/pixels/internal/realm/chat/history"
	"go.uber.org/fx"
)

const (
	// chatPath stores the chat administration base path.
	chatPath = "/api/admin/chat"
)

// Dependencies contains chat administration behavior.
type Dependencies struct {
	fx.In

	// Filter manages the global chat dictionary.
	Filter *chatfilter.Service
	// Bubbles manages bubble thresholds.
	Bubbles *bubble.Service
	// History reads bounded chat history.
	History *history.Service
}

// Register mounts protected chat administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	app.Get(chatPath+"/filters", listFilters(dependencies.Filter))
	app.Post(chatPath+"/filters", addFilter(dependencies.Filter))
	app.Delete(chatPath+"/filters/:word", removeFilter(dependencies.Filter))
	app.Get(chatPath+"/bubbles", listBubbles(dependencies.Bubbles))
	app.Put(chatPath+"/bubbles/:bubbleId", setBubble(dependencies.Bubbles))
	app.Delete(chatPath+"/bubbles/:bubbleId", deleteBubble(dependencies.Bubbles))
	app.Get("/api/admin/rooms/:id/chat/history", roomHistory(dependencies.History))
	app.Get("/api/admin/players/:playerId/chat/history", playerHistory(dependencies.History))
}
