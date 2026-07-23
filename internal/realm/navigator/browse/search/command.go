// Package search executes navigator room searches.
package search

import (
	"context"
	"strings"

	"github.com/niflaot/pixels/internal/command"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	navprojection "github.com/niflaot/pixels/internal/realm/navigator/browse/card"
	navevent "github.com/niflaot/pixels/internal/realm/navigator/browse/search/events/executed"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	navviewer "github.com/niflaot/pixels/internal/realm/navigator/viewer/live"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
	outsearch "github.com/niflaot/pixels/networking/outbound/navigator/browse/searchresult"
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
	// Groups attaches social-group room cards in one bounded batch.
	Groups GroupCards
	// Friends reads cached friendship identifiers.
	Friends FriendIDs
	// GroupRooms reads active memberships and headquarters.
	GroupRooms GroupRooms
	// RightRooms reads room identifiers with explicit rights.
	RightRooms RightRooms
	// Limit bounds ordinary result lists.
	Limit int
	// HistoryLimit bounds recent and frequent room history.
	HistoryLimit int
}

// GroupCards reads active room group bindings without roster queries.
type GroupCards interface {
	// RoomGroups returns active group metadata keyed by room identifier.
	RoomGroups(context.Context, []int64) (map[int64]grouprecord.Group, error)
}

// limit returns a normalized result bound.
func (handler Handler) limit() int {
	if handler.Limit <= 0 || handler.Limit > 100 {
		return DefaultLimit
	}
	return handler.Limit
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

	lists, count, err := handler.Result(ctx, player.ID(), envelope.Command.Code, envelope.Command.Data)
	if err != nil {
		return err
	}
	viewer := player.OpenNavigator()
	viewer.SetSearch(navviewer.LastSearch{Code: envelope.Command.Code, Query: envelope.Command.Data}, VisibleRoomIDs(lists))

	packet, err := outsearch.Encode(envelope.Command.Code, envelope.Command.Data, lists)
	if err != nil {
		return err
	}
	if err := envelope.Command.Handler.Send(ctx, packet); err != nil {
		return err
	}

	return handler.publish(ctx, player.ID(), envelope.Command.Code, envelope.Command.Data, count)
}

// Result builds navigator result lists for one player search.
func (handler Handler) Result(ctx context.Context, playerID int64, code string, data string) ([]outsearch.ResultList, int, error) {
	return handler.resultLists(ctx, playerID, code, data)
}

// VisibleRoomIDs extracts room ids from navigator result lists.
func VisibleRoomIDs(lists []outsearch.ResultList) []int64 {
	roomIDs := make([]int64, 0)
	for _, list := range lists {
		for _, room := range list.Rooms {
			if room.RoomID > 0 {
				roomIDs = append(roomIDs, int64(room.RoomID))
			}
		}
	}

	return roomIDs
}

// resultLists builds navigator result lists for the requested context.
func (handler Handler) resultLists(ctx context.Context, playerID int64, code string, data string) ([]outsearch.ResultList, int, error) {
	code = strings.TrimSpace(code)
	data = strings.TrimSpace(data)
	switch code {
	case "myworld_view", "my":
		return handler.myWorldLists(ctx, playerID)
	case "my_rooms":
		return handler.ownedLists(ctx, playerID)
	case "favorites", "my_favourite_rooms_search":
		return handler.favoriteLists(ctx, playerID)
	case "official_view":
		return handler.officialLists(ctx)
	case "history", "my_room_history_search":
		return handler.historyLists(ctx, playerID, false)
	case "frequent", "my_frequent_room_history_search":
		return handler.historyLists(ctx, playerID, true)
	case "rights", "my_room_rights_search":
		return handler.rightsLists(ctx, playerID)
	case "friends_now", "rooms_where_my_friends_are":
		return handler.friendCurrentLists(ctx, playerID)
	case "friends_rooms", "my_friends_rooms_search":
		return handler.friendOwnedLists(ctx, playerID)
	case "my_groups", "my_guild_bases_search":
		return handler.groupMembershipLists(ctx, playerID)
	case "groups":
		return handler.popularGroupLists(ctx)
	case "group_base", "guild_base_search":
		return handler.groupBaseLists(ctx, data)
	case "highest_score", "rooms_with_highest_score_search":
		return handler.highestScoreLists(ctx)
	case "recommended", "my_recommended_rooms":
		return handler.recommendedLists(ctx, playerID)
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
func (handler Handler) cards(ctx context.Context, rooms []roommodel.Room) ([]roomcard.Card, error) {
	roomIDs := make([]int64, len(rooms))
	for index := range rooms {
		roomIDs[index] = rooms[index].ID
	}
	groups := map[int64]grouprecord.Group{}
	tags := map[int64][]roommodel.Tag{}
	if finder, ok := handler.Rooms.(interface {
		ListTagsByRoomIDs(context.Context, []int64) (map[int64][]roommodel.Tag, error)
	}); ok {
		var err error
		tags, err = finder.ListTagsByRoomIDs(ctx, roomIDs)
		if err != nil {
			return nil, err
		}
	}
	if handler.Groups != nil {
		var err error
		groups, err = handler.Groups.RoomGroups(ctx, roomIDs)
		if err != nil {
			return nil, err
		}
	}
	cards := make([]roomcard.Card, 0, len(rooms))
	for index, room := range rooms {
		roomTags := tags[room.ID]
		values := make([]string, len(roomTags))
		for tagIndex := range roomTags {
			values[tagIndex] = roomTags[tagIndex].Value
		}
		card := navprojection.RoomCard(room, handler.userCount(room.ID), index+1, values)
		if group, found := groups[room.ID]; found {
			card.Group = &roomcard.Group{ID: int32(group.ID), Name: group.Name, Badge: group.BadgeCode}
		}
		cards = append(cards, card)
	}
	return cards, nil
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
