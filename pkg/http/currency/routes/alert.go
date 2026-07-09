package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	netconn "github.com/niflaot/pixels/networking/connection"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// sendMutationAlert sends an opt-in localized balance alert.
func sendMutationAlert(ctx *fiber.Ctx, dependencies Dependencies, action mutationAction, input mutationInput, balance int64) bool {
	if !input.request.Alert {
		return false
	}
	if dependencies.Translations == nil {
		return false
	}

	definition, found, err := currencyDefinition(ctx, dependencies, input.currencyType)
	if err != nil {
		dependencies.Log.Warn("currency admin alert translation lookup failed",
			zap.Int64("player_id", input.playerID),
			zap.Int32("currency_type", input.currencyType),
			zap.Error(err),
		)
		return false
	}
	if !found {
		return false
	}
	connection, found := playerConnection(dependencies, input.playerID)
	if !found {
		return false
	}

	locale := i18n.Locale(input.request.Locale)
	currencyName := dependencies.Translations.T(locale, i18n.Key("currency.name."+definition.Key))
	message := dependencies.Translations.T(locale, alertKey(action), i18n.Params{
		"amount":   strconv.FormatInt(input.request.Amount, 10),
		"balance":  strconv.FormatInt(balance, 10),
		"currency": currencyName,
	})
	packet, err := outalert.Encode(message)
	if err == nil {
		err = connection.Send(ctx.Context(), packet)
	}
	if err != nil {
		dependencies.Log.Warn("currency admin alert delivery failed",
			zap.Int64("player_id", input.playerID),
			zap.Int32("currency_type", input.currencyType),
			zap.String("action", string(action)),
			zap.Error(err),
		)
		return false
	}

	return true
}

// currencyDefinition finds one configured currency definition.
func currencyDefinition(ctx *fiber.Ctx, dependencies Dependencies, currencyType int32) (currencymodel.Definition, bool, error) {
	definitions, err := dependencies.Currencies.Types(ctx.Context())
	if err != nil {
		return currencymodel.Definition{}, false, err
	}
	for _, definition := range definitions {
		if definition.Type == currencyType {
			return definition, true, nil
		}
	}

	return currencymodel.Definition{}, false, nil
}

// playerConnection resolves one live player's connection.
func playerConnection(dependencies Dependencies, playerID int64) (netconn.Connection, bool) {
	if dependencies.Players == nil || dependencies.Connections == nil {
		return nil, false
	}
	player, found := dependencies.Players.Find(playerID)
	if !found {
		return nil, false
	}
	peer := player.Peer()

	return dependencies.Connections.Get(peer.ConnectionKind(), peer.ConnectionID())
}

// alertKey returns one stable localized admin mutation key.
func alertKey(action mutationAction) i18n.Key {
	return i18n.Key("admin.currency.alert." + string(action))
}
