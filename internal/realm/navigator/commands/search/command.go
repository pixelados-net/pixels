// Package search executes navigator room searches.
package search

import (
	"context"
	"strings"

	"github.com/niflaot/pixels/internal/command"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/commands/session"
	navevent "github.com/niflaot/pixels/internal/realm/navigator/events/searchexecuted"
	navprojection "github.com/niflaot/pixels/internal/realm/navigator/projection"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/service"
	navviewer "github.com/niflaot/pixels/internal/realm/navigator/viewer/live"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/networking/outbound/navigator/roomcard"
	outsearch "github.com/niflaot/pixels/networking/outbound/navigator/searchresult"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies the navigator search command.
	Name command.Name = "navigator.search"

	// DefaultLimit limits first navigator result groups.
	DefaultLimit int = 50
)

// Command searches navigator rooms.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
	// Code stores the navigator context or search code.
	Code string
	// Data stores the search query or filter.
	Data string
}

// Handler handles navigator search commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Navigator reads navigator persistence.
	Navigator navservice.Manager
	// Rooms reads room persistence.
	Rooms roomservice.Manager
	// Runtime reads active room occupancy.
	Runtime *roomlive.Registry
	// Events publishes navigator search events.
	Events bus.Publisher
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// MarshalLogObject writes safe debug command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Handler.ConnectionID))
	encoder.AddString("code", input.Code)
	encoder.AddString("data", input.Data)

	return nil
}

// Handle handles a navigator search command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, _, err := navsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}

	viewer := player.OpenNavigator()
	viewer.SetLastSearch(navviewer.LastSearch{Code: envelope.Command.Code, Query: envelope.Command.Data})
	lists, count, err := handler.resultLists(ctx, player.ID(), envelope.Command.Code, envelope.Command.Data)
	if err != nil {
		return err
	}

	packet, err := outsearch.Encode(envelope.Command.Code, envelope.Command.Data, lists)
	if err != nil {
		return err
	}
	if err := envelope.Command.Handler.Send(ctx, packet); err != nil {
		return err
	}

	return handler.publish(ctx, player.ID(), envelope.Command.Code, envelope.Command.Data, count)
}

// resultLists builds navigator result lists for the requested context.
func (handler Handler) resultLists(ctx context.Context, playerID int64, code string, data string) ([]outsearch.ResultList, int, error) {
	code = strings.TrimSpace(code)
	data = strings.TrimSpace(data)
	switch code {
	case "myworld_view", "my":
		return handler.myWorldLists(ctx, playerID)
	case "official_view":
		return handler.officialLists(ctx)
	case "roomads_view":
		return handler.eventLists(ctx)
	case "hotel_view", "popular":
		return handler.hotelLists(ctx, data)
	default:
		return handler.queryLists(ctx, code, data)
	}
}

// publish emits navigator search execution.
func (handler Handler) publish(ctx context.Context, playerID int64, code string, data string, count int) error {
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{Name: navevent.Name, Payload: navevent.Payload{PlayerID: playerID, Code: code, Query: data, Count: count}})
}

// cards maps room records to navigator cards.
func (handler Handler) cards(rooms []roommodel.Room) []roomcard.Card {
	cards := make([]roomcard.Card, 0, len(rooms))
	for index, room := range rooms {
		cards = append(cards, navprojection.RoomCard(room, handler.userCount(room.ID), index+1, nil))
	}

	return cards
}

// userCount returns live occupancy for a room.
func (handler Handler) userCount(roomID int64) int {
	if handler.Runtime == nil {
		return 0
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return 0
	}

	return active.Occupancy().Count
}
