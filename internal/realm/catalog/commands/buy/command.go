// Package buy executes catalog purchase requests.
package buy

import (
	"context"
	"errors"
	"fmt"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogprojection "github.com/niflaot/pixels/internal/realm/catalog/projection"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outsoldout "github.com/niflaot/pixels/networking/outbound/catalog/limited/soldout"
	outfailed "github.com/niflaot/pixels/networking/outbound/catalog/purchase/failed"
	outok "github.com/niflaot/pixels/networking/outbound/catalog/purchase/ok"
	outunavailable "github.com/niflaot/pixels/networking/outbound/catalog/purchase/unavailable"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies the catalog purchase command.
	Name command.Name = "catalog.item.buy"
)

// Command requests one catalog purchase.
type Command struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// PageID identifies the source page.
	PageID int64
	// OfferID identifies the purchased offer.
	OfferID int64
	// ExtraData stores client-supplied product data.
	ExtraData string
	// Amount stores the requested offer quantity.
	Amount int32
}

// Handler handles catalog purchase commands.
type Handler struct {
	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps authenticated connections to players.
	Bindings *binding.Registry
	// Catalog manages catalog purchases.
	Catalog catalogservice.Manager
	// Log records unexpected purchase failures.
	Log *zap.Logger
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// MarshalLogObject writes safe debug command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Connection.ConnectionID))
	encoder.AddInt64("page_id", input.PageID)
	encoder.AddInt64("offer_id", input.OfferID)
	encoder.AddInt32("amount", input.Amount)

	return nil
}

// Handle processes one catalog purchase and sends its protocol outcome.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := catalogsession.Player(envelope.Command.Connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if envelope.Command.Amount != 1 {
		return handler.sendUnavailable(ctx, envelope.Command.Connection)
	}
	_, items, err := handler.Catalog.Page(ctx, envelope.Command.PageID, catalogsession.DefaultRank, catalogsession.DefaultClub)
	if err != nil || !containsOffer(items, envelope.Command.OfferID) {
		if err == nil {
			err = catalogservice.ErrOfferNotFound
		}
		return handler.sendError(ctx, envelope.Command.Connection, envelope.Command.OfferID, err)
	}
	result, err := handler.Catalog.Purchase(ctx, catalogservice.PurchaseParams{
		PlayerID: player.ID(), CatalogItemID: envelope.Command.OfferID,
		Rank: catalogsession.DefaultRank, HasClub: catalogsession.DefaultClub,
	})
	if err != nil {
		return handler.sendError(ctx, envelope.Command.Connection, envelope.Command.OfferID, err)
	}
	definition, found, err := handler.Catalog.Definition(ctx, result.Item.DefinitionID)
	if err != nil || !found {
		if err == nil {
			err = fmt.Errorf("furniture definition %d not found", result.Item.DefinitionID)
		}
		return handler.sendError(ctx, envelope.Command.Connection, envelope.Command.OfferID, err)
	}
	if result.Item.IsLimited() {
		result.Item.LimitedSells++
	}
	mapped, err := catalogprojection.Offer(result.Item, definition)
	if err != nil {
		return handler.sendError(ctx, envelope.Command.Connection, envelope.Command.OfferID, err)
	}
	itemIDs := make([]int64, 0, len(result.GrantedItems))
	for _, item := range result.GrantedItems {
		itemIDs = append(itemIDs, item.ID)
	}
	packet, err := outunseen.EncodeOwned(itemIDs)
	if err != nil {
		return err
	}
	if err := envelope.Command.Connection.Send(ctx, packet); err != nil {
		return err
	}
	packet, err = outok.Encode(mapped)
	if err != nil {
		return err
	}
	if err := envelope.Command.Connection.Send(ctx, packet); err != nil {
		return err
	}
	refresh, err := outrefresh.Encode()
	if err != nil {
		return err
	}

	return envelope.Command.Connection.Send(ctx, refresh)
}

// containsOffer reports whether a page response contains one offer id.
func containsOffer(items []catalogmodel.Item, offerID int64) bool {
	for _, item := range items {
		if item.ID == offerID {
			return true
		}
	}

	return false
}

// sendError maps a catalog service failure to its protocol result.
func (handler Handler) sendError(ctx context.Context, connection netconn.Context, offerID int64, err error) error {
	if errors.Is(err, catalogservice.ErrLimitedSoldOut) {
		packet, encodeErr := outsoldout.Encode()
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, packet)
	}
	if errors.Is(err, catalogservice.ErrOfferNotFound) || errors.Is(err, catalogservice.ErrOfferNotVisible) ||
		errors.Is(err, catalogservice.ErrOfferDisabled) || errors.Is(err, catalogservice.ErrPageNotFound) ||
		errors.Is(err, currencyservice.ErrInsufficientBalance) {
		return handler.sendUnavailable(ctx, connection)
	}
	if handler.Log != nil {
		handler.Log.Error("catalog purchase failed", zap.Int64("offer_id", offerID), zap.Error(err))
	}
	packet, encodeErr := outfailed.Encode(outfailed.CodeServer)
	if encodeErr != nil {
		return encodeErr
	}

	return connection.Send(ctx, packet)
}

// sendUnavailable sends an illegal purchase response.
func (handler Handler) sendUnavailable(ctx context.Context, connection netconn.Context) error {
	packet, err := outunavailable.Encode(outunavailable.CodeIllegal)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
