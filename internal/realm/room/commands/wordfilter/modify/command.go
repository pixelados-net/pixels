// Package modify mutates room word filters.
package modify

import (
	"context"
	"strings"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/commands/control"
	wordfiltermodified "github.com/niflaot/pixels/internal/realm/room/events/wordfiltermodified"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies the room word filter modify command.
	Name command.Name = "room.word_filter.modify"
)

// Command contains one room filter mutation.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the room.
	RoomID int64
	// Add reports whether to add instead of remove.
	Add bool
	// Word stores the filter word.
	Word string
}

// Handler handles room filter mutations.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Filters manages room filter words.
	Filters FilterMutator
	// Events publishes committed filter events.
	Events bus.Publisher
}

// FilterMutator mutates room filter words.
type FilterMutator interface {
	// Add adds one room filter word.
	Add(context.Context, int64, int64, string) error
	// Remove removes one room filter word.
	Remove(context.Context, int64, int64, string) error
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle mutates one room filter word.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	input := envelope.Command
	player, roomID, err := control.Actor(input.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err = control.MatchRoom(roomID, input.RoomID); err != nil {
		return err
	}
	if input.Add {
		err = handler.Filters.Add(ctx, roomID, player.ID(), input.Word)
	} else {
		err = handler.Filters.Remove(ctx, roomID, player.ID(), input.Word)
	}
	if err != nil {
		return err
	}
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{Name: wordfiltermodified.Name, Payload: wordfiltermodified.Payload{RoomID: roomID, ActorID: player.ID(), Added: input.Add, Word: strings.ToLower(strings.TrimSpace(input.Word))}})
}
