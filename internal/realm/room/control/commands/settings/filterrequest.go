package settings

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outlist "github.com/niflaot/pixels/networking/outbound/room/wordfilter/list"
)

const (
	// FilterRequestName identifies the room word filter request command.
	FilterRequestName command.Name = "room.word_filter.request"
)

// FilterRequestCommand requests room filter words.
type FilterRequestCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the room.
	RoomID int64
}

// FilterRequestHandler handles room filter requests.
type FilterRequestHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rooms reads room metadata.
	Rooms FilterRoomFinder
	// Authorize resolves settings capability.
	Authorize *roomsettings.Authorizer
	// Filters manages room filter words.
	Filters FilterLister
}

// FilterRoomFinder reads room metadata.
type FilterRoomFinder interface {
	// FindByID finds one room.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
}

// FilterLister lists room filter words.
type FilterLister interface {
	// List lists room filter words.
	List(context.Context, int64) ([]string, error)
}

// CommandName returns the stable command name.
func (FilterRequestCommand) CommandName() command.Name { return FilterRequestName }

// Handle sends current room filter words.
func (handler FilterRequestHandler) Handle(ctx context.Context, envelope command.Envelope[FilterRequestCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err = control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}
	room, found, err := handler.Rooms.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if !found {
		return roomservice.ErrRoomNotFound
	}
	if err = handler.Authorize.Authorize(ctx, room, player.ID()); err != nil {
		return err
	}
	words, err := handler.Filters.List(ctx, roomID)
	if err != nil {
		return err
	}
	packet, err := outlist.Encode(words)
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}
