// Package flatcats sends available room categories to navigator clients.
package flatcats

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outflatcats "github.com/niflaot/pixels/networking/outbound/navigator/browse/flatcats"
)

const (
	// Name identifies the navigator flat categories command.
	Name command.Name = "navigator.flat_cats"
)

// CategoryReader reads visible room categories.
type CategoryReader interface {
	// ListCategories lists room categories.
	ListCategories(ctx context.Context) ([]roommodel.Category, error)
}

// Command sends navigator flat categories.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
}

// Handler handles flat category requests.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Categories reads room category records.
	Categories CategoryReader
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a flat category command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	if _, _, err := navsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players); err != nil {
		return err
	}

	categories, err := handler.Categories.ListCategories(ctx)
	if err != nil {
		return err
	}

	packet, err := outflatcats.Encode(flatCategories(categories))
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}

// flatCategories maps room category records.
func flatCategories(categories []roommodel.Category) []outflatcats.Category {
	results := make([]outflatcats.Category, 0, len(categories))
	for _, category := range categories {
		results = append(results, outflatcats.Category{
			ID:                   int32(category.ID),
			Name:                 category.Caption,
			Visible:              category.Visible,
			Automatic:            category.Automatic,
			AutomaticCategoryKey: category.AutomaticKey,
			GlobalCategoryKey:    category.GlobalKey,
			StaffOnly:            category.StaffOnly,
		})
	}

	return results
}
