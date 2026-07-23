// Package pages executes catalog index requests.
package pages

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	catalogprojection "github.com/niflaot/pixels/internal/realm/catalog/projection"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outpages "github.com/niflaot/pixels/networking/outbound/catalog/pages"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies the catalog pages command.
	Name command.Name = "catalog.pages.request"
)

// Command requests the visible catalog tree.
type Command struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// Mode stores the requested catalog mode.
	Mode string
}

// Handler handles catalog page tree requests.
type Handler struct {
	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps authenticated connections to players.
	Bindings *binding.Registry
	// Catalog reads cached catalog data.
	Catalog catalogservice.Reader
	// Translations localizes catalog page names.
	Translations i18n.Translator
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// MarshalLogObject writes safe debug command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Connection.ConnectionID))
	encoder.AddString("mode", input.Mode)

	return nil
}

// Handle sends the visible catalog tree.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := catalogsession.Player(envelope.Command.Connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	viewer := player.OpenCatalog()
	viewer.SetMode(envelope.Command.Mode)
	pages, err := handler.Catalog.Pages(ctx, player.ID(), catalogsession.HasClub(player))
	if err != nil {
		return err
	}
	nodes, err := catalogprojection.PageTree(pages, handler.Translations)
	if err != nil {
		return err
	}
	newAdditions := false
	if novelty, ok := handler.Catalog.(catalogservice.NoveltyManager); ok {
		newAdditions, err = novelty.NewAdditionsAvailable(ctx, player.ID())
		if err != nil {
			return err
		}
	}
	packet, err := outpages.Encode(nodes, viewer.Mode(), newAdditions)
	if err != nil {
		return err
	}

	return envelope.Command.Connection.Send(ctx, packet)
}
