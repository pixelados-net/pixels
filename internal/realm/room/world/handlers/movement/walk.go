// Package movement adapts room unit movement packets into commands.
package movement

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	lookcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/look"
	walkcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/walk"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inwalk "github.com/niflaot/pixels/networking/inbound/room/entities/walk"
	"go.uber.org/zap"
)

// NewWalk creates a room walk packet handler.
func NewWalk(handler walkcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inwalk.Decode(packet)
		if err != nil {
			return err
		}

		err = dispatcher.Dispatch(context.Background(), command.Envelope[walkcmd.Command]{
			Command:  walkcmd.Command{Handler: connection, X: int(payload.X), Y: int(payload.Y)},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
		if isStalePresenceError(err) {
			if log != nil {
				log.Debug("room walk ignored after presence ended", zap.Error(err), zap.String("connection_id", string(connection.ConnectionID)))
			}
			return nil
		}

		return err
	}
}

// isStalePresenceError reports a room packet received after presence teardown.
func isStalePresenceError(err error) bool {
	return errors.Is(err, walkcmd.ErrPlayerNotInRoom) || errors.Is(err, lookcmd.ErrPlayerNotInRoom) || errors.Is(err, roomlive.ErrRoomNotFound)
}

// RegisterWalk adds the room walk handler to a registry.
func RegisterWalk(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inwalk.Header, handler)
}
