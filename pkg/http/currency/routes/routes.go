// Package routes contains protected currency administration routes.
package routes

import (
	"github.com/gofiber/fiber/v2"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	// playerBasePath stores player currency route prefixes.
	playerBasePath = "/api/admin/players"

	// catalogPath stores the currency catalog route.
	catalogPath = "/api/admin/currencies/types"
)

// Dependencies contains currency administration collaborators.
type Dependencies struct {
	fx.In

	// Finder validates persistent player identities.
	Finder playerservice.Finder

	// Players stores live player compositions.
	Players *playerlive.Registry

	// Connections stores active protocol connections.
	Connections *netconn.Registry

	// Currencies manages persistent currency balances.
	Currencies currencyservice.Manager

	// Translations resolves optional player alerts.
	Translations i18n.Translator

	// Log records optional alert delivery failures.
	Log *zap.Logger
}

// Register registers protected currency administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	if dependencies.Log == nil {
		dependencies.Log = zap.NewNop()
	}

	app.Get(playerBasePath+"/:id/currencies", walletHandler(dependencies))
	app.Post(playerBasePath+"/:id/currencies/:type/grant", mutationHandler(grantAction, dependencies))
	app.Post(playerBasePath+"/:id/currencies/:type/deduct", mutationHandler(deductAction, dependencies))
	app.Post(playerBasePath+"/:id/currencies/:type/set", mutationHandler(setAction, dependencies))
	app.Get(catalogPath, typesHandler(dependencies))
}
