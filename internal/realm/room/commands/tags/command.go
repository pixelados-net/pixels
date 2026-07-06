// Package tags sends room tag results.
package tags

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomsession "github.com/niflaot/pixels/internal/realm/room/commands/session"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outtags "github.com/niflaot/pixels/networking/outbound/room/tags"
)

const (
	// Name identifies the room tags command.
	Name command.Name = "room.tags"
)

// Command sends room tags.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
}

// Handler handles room tags commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Rooms reads room persistence.
	Rooms roomservice.Manager
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a room tags command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := roomsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}

	roomID, found := player.CurrentRoom()
	if !found {
		return handler.send(ctx, envelope.Command.Handler, nil)
	}

	tags, err := handler.Rooms.ListTags(ctx, roomID)
	if err != nil {
		return err
	}

	return handler.send(ctx, envelope.Command.Handler, tagEntries(tags))
}

// send sends room tag entries.
func (handler Handler) send(ctx context.Context, connection netconn.Context, entries []outtags.Entry) error {
	packet, err := outtags.Encode(entries)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
