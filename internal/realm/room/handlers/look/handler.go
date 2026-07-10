// Package look contains the room look packet handler.
package look

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	lookcmd "github.com/niflaot/pixels/internal/realm/room/commands/look"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlook "github.com/niflaot/pixels/networking/inbound/room/entities/look"
	"go.uber.org/zap"
)

// New creates a room look packet handler.
func New(handler lookcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inlook.Decode(packet)
		if err != nil {
			return err
		}

		err = dispatcher.Dispatch(context.Background(), command.Envelope[lookcmd.Command]{
			Command:  lookcmd.Command{Handler: connection, X: int(payload.X), Y: int(payload.Y)},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
		if isStalePresenceError(err) {
			if log != nil {
				log.Debug("room look ignored after presence ended", zap.Error(err), zap.String("connection_id", string(connection.ConnectionID)))
			}
			return nil
		}

		return err
	}
}

// isStalePresenceError reports a room packet received after presence teardown.
func isStalePresenceError(err error) bool {
	return errors.Is(err, lookcmd.ErrPlayerNotInRoom) || errors.Is(err, roomlive.ErrRoomNotFound)
}

// Register adds the room look handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inlook.Header, handler)
}
