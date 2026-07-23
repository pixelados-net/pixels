// Package cancreate checks whether a player can create rooms.
package cancreate

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcancreate "github.com/niflaot/pixels/networking/outbound/navigator/create/cancreate"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies the navigator can-create command.
	Name command.Name = "navigator.can_create"

	// ResultAllowed tells the client room creation is allowed.
	ResultAllowed int32 = 0
	// ResultLimitReached tells the client its room ownership limit was reached.
	ResultLimitReached int32 = 1

	// RoomLimit stores the initial per-player room limit.
	RoomLimit int32 = roomservice.MaxRoomsPerPlayer
)

// Command checks room creation permission.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
}

// Handler handles room creation preflight commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Rooms reads player-owned rooms.
	Rooms roomservice.OwnerLister
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

// Handle handles a room creation preflight command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, _, err := navsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	rooms, err := handler.Rooms.ListByOwner(ctx, player.ID())
	if err != nil {
		return err
	}
	result := ResultAllowed
	if len(rooms) >= int(RoomLimit) {
		result = ResultLimitReached
	}

	packet, err := outcancreate.Encode(result, RoomLimit)
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}
