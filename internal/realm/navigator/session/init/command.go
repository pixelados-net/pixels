// Package init initializes the live navigator viewer.
package init

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	navevent "github.com/niflaot/pixels/internal/realm/navigator/session/events/initialized"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outmetadata "github.com/niflaot/pixels/networking/outbound/navigator/session/metadata"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies the navigator init command.
	Name command.Name = "navigator.init"

	// FavoriteLimit is the default room favorite limit.
	FavoriteLimit int32 = 30
)

// Command initializes a player's navigator viewer.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
}

// Handler handles navigator init commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Navigator reads navigator persistence.
	Navigator navservice.Manager
	// Rooms reads room persistence.
	Rooms roomservice.Manager
	// Events publishes navigator lifecycle events.
	Events bus.Publisher
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// MarshalLogObject writes safe debug command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Handler.ConnectionID))

	return nil
}

// Handle handles a navigator init command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, _, err := navsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}

	player.OpenNavigator()
	if err := handler.sendInitialPackets(ctx, envelope.Command.Handler, player.ID()); err != nil {
		return err
	}

	return handler.publish(ctx, player.ID())
}

// sendInitialPackets sends the navigator bootstrap packet set.
func (handler Handler) sendInitialPackets(ctx context.Context, connection netconn.Context, playerID int64) error {
	saved, err := handler.Navigator.ListSavedSearches(ctx, playerID)
	if err != nil {
		return err
	}
	preference, err := handler.Navigator.Preference(ctx, playerID)
	if err != nil {
		return err
	}
	lifted, err := handler.Navigator.ListLiftedRooms(ctx)
	if err != nil {
		return err
	}
	favorites, err := handler.Navigator.ListFavoriteRoomIDs(ctx, playerID)
	if err != nil {
		return err
	}
	categoryPreferences, err := handler.Navigator.ListCategoryPreferences(ctx, playerID)
	if err != nil {
		return err
	}
	categories, err := handler.Rooms.ListCategories(ctx)
	if err != nil {
		return err
	}

	packets, err := initialPackets(saved, preference, lifted, favorites, categoryPreferences, categories)
	if err != nil {
		return err
	}

	return sendPackets(ctx, connection, packets)
}

// publish emits navigator initialization.
func (handler Handler) publish(ctx context.Context, playerID int64) error {
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{Name: navevent.Name, Payload: navevent.Payload{PlayerID: playerID}})
}

// metadataContexts returns top-level Nitro navigator contexts.
func metadataContexts(saved []navmodel.SavedSearch) []outmetadata.Context {
	return []outmetadata.Context{
		{Code: "official_view", SavedSearches: metadataSearches(saved, "official_view")},
		{Code: "hotel_view", SavedSearches: metadataSearches(saved, "hotel_view")},
		{Code: "roomads_view", SavedSearches: metadataSearches(saved, "roomads_view")},
		{Code: "myworld_view", SavedSearches: metadataSearches(saved, "myworld_view")},
	}
}
