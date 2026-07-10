// Package page executes catalog page requests.
package page

import (
	"context"
	"fmt"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	catalogprojection "github.com/niflaot/pixels/internal/realm/catalog/projection"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
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
	page, items, err := handler.Catalog.Page(ctx, envelope.Command.PageID, catalogsession.DefaultRank, catalogsession.DefaultClub)
	if err != nil {
		return err
	}
	offers := make([]offer.Offer, 0, len(items))
	for _, item := range items {
		definition, found, err := handler.Catalog.Definition(ctx, item.DefinitionID)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("catalog item %d furniture definition %d not found", item.ID, item.DefinitionID)
		}
		mapped, err := catalogprojection.Offer(item, definition)
		if err != nil {
			return fmt.Errorf("map catalog item %d: %w", item.ID, err)
		}
		offers = append(offers, mapped)
	}
	viewer.SetPage(page.ID)
	pageText := "catalog.page." + page.Name
	if handler.Translations != nil {
		pageText = handler.Translations.Default(i18n.Key(pageText))
	}
	packet, err := outpage.Encode(int32(page.ID), viewer.Mode(), page.Layout,
		outpage.Localization{Images: []string{"", "", ""}, Texts: []string{pageText, "", ""}}, offers, envelope.Command.OfferID)
	if err != nil {
		return err
	}

	return envelope.Command.Connection.Send(ctx, packet)
}
