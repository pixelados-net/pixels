// Package page executes catalog page requests.
package page

import (
	"context"
	"fmt"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogprojection "github.com/niflaot/pixels/internal/realm/catalog/projection"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/networking/outbound/catalog/offer"
	outpage "github.com/niflaot/pixels/networking/outbound/catalog/page"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies the catalog page command.
	Name command.Name = "catalog.page.request"
)

// Command requests one catalog page.
type Command struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// PageID identifies the requested page.
	PageID int64
	// OfferID identifies an optional highlighted offer.
	OfferID int32
	// Mode stores the requested catalog mode.
	Mode string
}

// Handler handles catalog page requests.
type Handler struct {
	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps authenticated connections to players.
	Bindings *binding.Registry
	// Catalog reads cached catalog data.
	Catalog catalogservice.Reader
	// Translations localizes catalog page text.
	Translations i18n.Translator
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// MarshalLogObject writes safe debug command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Connection.ConnectionID))
	encoder.AddInt64("page_id", input.PageID)
	encoder.AddInt32("offer_id", input.OfferID)

	return nil
}

// Handle sends one visible catalog page.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := catalogsession.Player(envelope.Command.Connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	viewer := player.OpenCatalog()
	viewer.SetMode(envelope.Command.Mode)
	page, items, err := handler.Catalog.Page(ctx, envelope.Command.PageID, player.ID(), catalogsession.HasClub(player))
	if err != nil {
		return err
	}
	offers := make([]offer.Offer, 0, len(items))
	for _, item := range items {
		products := products(handler.Catalog, ctx, item)
		if len(products) == 0 {
			products = []catalogmodel.Product{{DefinitionID: item.DefinitionID, Quantity: item.Amount}}
		}
		definitions := make(map[int64]furnituremodel.Definition, len(products))
		for _, product := range products {
			definition, found, findErr := handler.Catalog.Definition(ctx, product.DefinitionID)
			if findErr != nil {
				return findErr
			}
			if !found {
				return fmt.Errorf("catalog item %d furniture definition %d not found", item.ID, product.DefinitionID)
			}
			definitions[product.DefinitionID] = definition
		}
		mapped, err := catalogprojection.OfferProducts(item, products, definitions)
		if err != nil {
			return fmt.Errorf("map catalog item %d: %w", item.ID, err)
		}
		// Room bundles are server-defined offers and their names are not part of
		// Nitro's static product-data/localization files. Send the translated
		// label as the fallback localization id so the client never exposes the
		// internal catalog.item.* key in the room selector.
		if item.IsRoomBundle() && handler.Translations != nil {
			mapped.LocalizationID = handler.Translations.Default(i18n.Key("catalog.item." + item.Name))
		}
		offers = append(offers, mapped)
	}
	viewer.SetPage(page.ID)
	pageText := "catalog.page." + page.Name
	if handler.Translations != nil {
		pageText = handler.Translations.Default(i18n.Key(pageText))
	}
	texts := []string{pageText, "", ""}
	if page.Layout == "room_bundle" && handler.Translations != nil {
		texts[1] = handler.Translations.Default(i18n.Key("catalog.page." + page.Name + ".teaser"))
		texts[2] = handler.Translations.Default(i18n.Key("catalog.page." + page.Name + ".details"))
	}
	packet, err := outpage.Encode(int32(page.ID), viewer.Mode(), page.Layout,
		outpage.Localization{Images: []string{"", "", ""}, Texts: texts}, offers, envelope.Command.OfferID)
	if err != nil {
		return err
	}

	return envelope.Command.Connection.Send(ctx, packet)
}

// products resolves optional bundle products with a legacy single-product fallback.
func products(reader catalogservice.Reader, ctx context.Context, item catalogmodel.Item) []catalogmodel.Product {
	bundles, ok := reader.(catalogservice.BundleReader)
	if ok {
		products := bundles.Products(ctx, item.ID)
		if len(products) != 0 {
			return products
		}
	}

	return []catalogmodel.Product{{DefinitionID: item.DefinitionID, Quantity: item.Amount}}
}
