// Package clientconfig serves data-driven Nitro client configuration.
package clientconfig

import (
	"github.com/gofiber/fiber/v2"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	"github.com/niflaot/pixels/pkg/i18n"
)

const (
	// UIConfigPath is the public Nitro configuration extension path.
	UIConfigPath = "/client/ui-config.json"

	// ExternalTextsPath is the public localized Nitro text extension path.
	ExternalTextsPath = "/client/texts/:locale/ExternalTexts.json"
)

// Register registers public Nitro currency configuration routes.
func Register(app *fiber.App, currencies currencyservice.Reader, translations i18n.Translator) {
	app.Get(UIConfigPath, uiConfigHandler(currencies))
	app.Get(ExternalTextsPath, externalTextsHandler(currencies, translations))
}
