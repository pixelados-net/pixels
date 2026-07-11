// Package request sends room word filters.
package request

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/commands/control"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/settings"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outlist "github.com/niflaot/pixels/networking/outbound/room/wordfilter/list"
)

const (
	// Name identifies the room word filter request command.
	Name command.Name = "room.word_filter.request"
)

// Command requests room filter words.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the room.
	RoomID int64
}

// Handler handles room filter requests.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rooms reads room metadata.
	Rooms RoomFinder
	// Authorize resolves settings capability.
	Authorize *roomsettings.Authorizer
	// Filters manages room filter words.
	Filters FilterLister
}

// RoomFinder reads room metadata.
type RoomFinder interface {
	// FindByID finds one room.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
}

// FilterLister lists room filter words.
type FilterLister interface {
	// List lists room filter words.
	List(context.Context, int64) ([]string, error)
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle sends current room filter words.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
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
