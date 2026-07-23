// Package handlers adapts crafting packets to recipe workflows.
package handlers

import (
	"context"
	"errors"

	craftingrecipe "github.com/niflaot/pixels/internal/realm/crafting/recipe"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnitureprojection "github.com/niflaot/pixels/internal/realm/furniture/projection"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inaltar "github.com/niflaot/pixels/networking/inbound/crafting/altar/products"
	incraft "github.com/niflaot/pixels/networking/inbound/crafting/craft"
	inrecipe "github.com/niflaot/pixels/networking/inbound/crafting/recipe/get"
	inhint "github.com/niflaot/pixels/networking/inbound/crafting/recipe/hint"
	insecret "github.com/niflaot/pixels/networking/inbound/crafting/secret/craft"
	outsoldout "github.com/niflaot/pixels/networking/outbound/catalog/limited/soldout"
	outproducts "github.com/niflaot/pixels/networking/outbound/crafting/altar/products"
	outcraft "github.com/niflaot/pixels/networking/outbound/crafting/craft"
	outrecipe "github.com/niflaot/pixels/networking/outbound/crafting/recipe/get"
	outhint "github.com/niflaot/pixels/networking/outbound/crafting/recipe/hint"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler handles every altar crafting packet.
type Handler struct {
	Service      *craftingrecipe.Service
	Players      *playerlive.Registry
	Bindings     *binding.Registry
	Translations i18n.Translator
}

// New creates a grouped recipe packet handler.
func New(service *craftingrecipe.Service, players *playerlive.Registry, bindings *binding.Registry, translations i18n.Translator) *Handler {
	return &Handler{Service: service, Players: players, Bindings: bindings, Translations: translations}
}

// Handle decodes and executes one crafting request.
func (handler *Handler) Handle(connection netconn.Context, packet codec.Packet) error {
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return handler.reject(connection, craftingrecord.ErrAltarNotFound)
	}
	ctx := context.Background()
	switch packet.Header {
	case inaltar.Header:
		payload, decodeErr := inaltar.Decode(packet)
		if decodeErr != nil {
			return decodeErr
		}
		recipes, openErr := handler.Service.Open(ctx, player.ID(), roomID, int64(payload.AltarItemID))
		if openErr != nil {
			return handler.reject(connection, openErr)
		}
		response, encodeErr := outproducts.Encode(recipes)
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, response)
	case inrecipe.Header:
		payload, decodeErr := inrecipe.Decode(packet)
		if decodeErr != nil {
			return decodeErr
		}
		recipe, findErr := handler.Service.Recipe(ctx, player.ID(), payload.RecipeName)
		if findErr != nil {
			return handler.reject(connection, findErr)
		}
		response, encodeErr := outrecipe.Encode(recipe.Ingredients)
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, response)
	case incraft.Header:
		payload, decodeErr := incraft.Decode(packet)
		if decodeErr != nil {
			return decodeErr
		}
		result, craftErr := handler.Service.Craft(ctx, player.ID(), roomID, int64(payload.AltarItemID), payload.RecipeName)
		return handler.craftResult(ctx, connection, result, craftErr)
	case insecret.Header:
		payload, decodeErr := insecret.Decode(packet)
		if decodeErr != nil {
			return decodeErr
		}
		result, craftErr := handler.Service.CraftSecret(ctx, player.ID(), roomID, payload.AltarItemID, payload.ItemIDs)
		return handler.craftResult(ctx, connection, result, craftErr)
	case inhint.Header:
		payload, decodeErr := inhint.Decode(packet)
		if decodeErr != nil {
			return decodeErr
		}
		count, exact, hintErr := handler.Service.Hint(ctx, player.ID(), roomID, payload.AltarItemID, payload.ItemIDs)
		if hintErr != nil {
			return handler.reject(connection, hintErr)
		}
		response, encodeErr := outhint.Encode(count, exact)
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, response)
	default:
		return codec.ErrUnexpectedHeader
	}
}

func (handler *Handler) craftResult(ctx context.Context, connection netconn.Context, result craftingrecipe.Result, err error) error {
	if err != nil {
		if errors.Is(err, craftingrecord.ErrRecipeSoldOut) {
			packet, encodeErr := outsoldout.Encode()
			if encodeErr != nil {
				return encodeErr
			}
			return connection.Send(ctx, packet)
		}
		return handler.reject(connection, err)
	}
	if err = furnitureprojection.Inventory(ctx, connection, result.Removed, result.Granted, result.Definition); err != nil {
		return err
	}
	packet, err := outcraft.Encode(true, outcraft.WithProduct(result.Recipe.Name, result.Recipe.RewardName))
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

func (handler *Handler) reject(connection netconn.Context, cause error) error {
	if !craftingrecipe.Expected(cause) {
		return cause
	}
	key := i18n.Key("crafting.error.unavailable")
	if errors.Is(cause, craftingrecord.ErrIngredients) || errors.Is(cause, craftingrecord.ErrItemUnavailable) {
		key = "crafting.error.ingredients"
	}
	message := string(key)
	if handler.Translations != nil {
		message = handler.Translations.Default(key)
	}
	packet, err := outalert.Encode(message)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), packet)
}

// Register adds all recipe headers to a connection registry.
func Register(registry *netconn.HandlerRegistry, handler *Handler) {
	for _, header := range []uint16{inaltar.Header, inrecipe.Header, incraft.Header, insecret.Header, inhint.Header} {
		_ = registry.Register(header, handler.Handle)
	}
}
