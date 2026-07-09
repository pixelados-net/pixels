package clientconfig

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	"github.com/niflaot/pixels/pkg/i18n"
)

// UIConfigResponse contains Nitro currency display configuration.
type UIConfigResponse struct {
	// CurrencyTypes stores every configured protocol currency type.
	CurrencyTypes []int32 `json:"system.currency.types"`
}

// ExternalTextsResponse stores Nitro text keys and localized values.
type ExternalTextsResponse map[string]string

// uiConfigHandler serves configured Nitro currency types.
func uiConfigHandler(currencies currencyservice.Reader) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		definitions, err := currencies.Types(ctx.Context())
		if err != nil {
			return err
		}
		types := make([]int32, 0, len(definitions))
		for _, definition := range definitions {
			types = append(types, definition.Type)
		}

		allowClientConfigOrigin(ctx)

		return ctx.JSON(UIConfigResponse{CurrencyTypes: types})
	}
}

// externalTextsHandler serves localized Nitro currency names.
func externalTextsHandler(currencies currencyservice.Reader, translations i18n.Translator) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		definitions, err := currencies.Types(ctx.Context())
		if err != nil {
			return err
		}
		locale := i18n.Locale(ctx.Params("locale"))
		texts := make(ExternalTextsResponse, len(definitions))
		for _, definition := range definitions {
			texts["purse.seasonal.currency."+currencyTypeString(definition.Type)] = translations.T(
				locale,
				i18n.Key("currency.name."+definition.Key),
			)
		}

		allowClientConfigOrigin(ctx)

		return ctx.JSON(texts)
	}
}

// allowClientConfigOrigin permits Nitro clients hosted on a separate origin.
func allowClientConfigOrigin(ctx *fiber.Ctx) {
	ctx.Set(fiber.HeaderAccessControlAllowOrigin, "*")
}

// currencyTypeString formats a protocol currency type.
func currencyTypeString(currencyType int32) string {
	return strconv.FormatInt(int64(currencyType), 10)
}
