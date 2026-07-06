// Package cancreate checks whether a player can create rooms.
package cancreate

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/commands/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcancreate "github.com/niflaot/pixels/networking/outbound/navigator/cancreate"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies the navigator can-create command.
	Name command.Name = "navigator.can_create"

	// ResultAllowed tells the client room creation is allowed.
	ResultAllowed int32 = 0

	// RoomLimit stores the initial per-player room limit.
	RoomLimit int32 = 100
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
	if _, _, err := navsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players); err != nil {
		return err
	}

	packet, err := outcancreate.Encode(ResultAllowed, RoomLimit)
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}
