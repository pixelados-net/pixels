package clientconfig

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/pkg/i18n"
)

// UIConfigResponse contains Nitro currency display configuration.
type UIConfigResponse struct {
	// CurrencyTypes stores every configured protocol currency type.
	CurrencyTypes []int32 `json:"system.currency.types"`
	// RoomModels stores server-enabled Nitro room creator models.
	RoomModels []RoomModel `json:"navigator.room.models"`
}

// RoomModel describes one Nitro room creator model.
type RoomModel struct {
	// ClubLevel stores the client entitlement level.
	ClubLevel int `json:"clubLevel"`
	// TileSize stores the displayed usable tile count.
	TileSize int `json:"tileSize"`
	// Name stores the model suffix expected by Nitro.
	Name string `json:"name"`
}

// ExternalTextsResponse stores Nitro text keys and localized values.
type ExternalTextsResponse map[string]string

// uiConfigHandler serves configured Nitro currency types.
func uiConfigHandler(currencies currencyservice.Reader, layouts roomlayout.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		definitions, err := currencies.Types(ctx.Context())
		if err != nil {
			return err
		}
		types := make([]int32, 0, len(definitions))
		for _, definition := range definitions {
			types = append(types, definition.Type)
		}
		available, err := layouts.List(ctx.Context())
		if err != nil {
			return err
		}
		models := make([]RoomModel, 0, len(available))
		for _, roomModel := range available {
			if !roomModel.Enabled {
				continue
			}
			models = append(models, RoomModel{ClubLevel: 0, TileSize: roomModel.TileSize, Name: strings.TrimPrefix(roomModel.Name, "model_")})
		}

		allowClientConfigOrigin(ctx)

		return ctx.JSON(UIConfigResponse{CurrencyTypes: types, RoomModels: models})
	}
}

// externalTextsHandler serves localized Nitro external texts.
func externalTextsHandler(currencies currencyservice.Reader, translations i18n.Translator) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		definitions, err := currencies.Types(ctx.Context())
		if err != nil {
			return err
		}
		locale := i18n.Locale(ctx.Params("locale"))
		entries := translations.Entries(locale)
		texts := make(ExternalTextsResponse, len(entries)+len(definitions))
		for key, value := range entries {
			texts[string(key)] = value
		}
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
