// Package clientconfig serves data-driven Nitro client configuration.
package clientconfig

import (
	"github.com/gofiber/fiber/v2"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/pkg/i18n"
)

const (
	// UIConfigPath is the public Nitro configuration extension path.
	UIConfigPath = "/client/ui-config.json"

	// ExternalTextsPath is the public localized Nitro text extension path.
	ExternalTextsPath = "/client/texts/:locale/ExternalTexts.json"
)

// Register registers public data-driven Nitro configuration routes.
func Register(app *fiber.App, currencies currencyservice.Reader, layouts roomlayout.Manager, translations i18n.Translator) {
	app.Get(UIConfigPath, uiConfigHandler(currencies, layouts))
	app.Get(ExternalTextsPath, externalTextsHandler(currencies, translations))
}
