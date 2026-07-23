package floorplan

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	floorplancmd "github.com/niflaot/pixels/internal/realm/room/control/commands/floorplan"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inblocked "github.com/niflaot/pixels/networking/inbound/room/floorplan/blockedtiles"
	"go.uber.org/zap"
)

// NewBlockedTiles creates an occupied floor plan tiles packet handler.
func NewBlockedTiles(handler floorplancmd.BlockedTilesHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := inblocked.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[floorplancmd.BlockedTilesCommand]{Command: floorplancmd.BlockedTilesCommand{Handler: connection}})
	}
}

// RegisterBlockedTiles adds the occupied floor plan tiles handler.
func RegisterBlockedTiles(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inblocked.Header, handler)
}
