package floorplan

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	floorplancmd "github.com/niflaot/pixels/internal/realm/room/control/commands/floorplan"
	domain "github.com/niflaot/pixels/internal/realm/room/control/floorplan"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	insave "github.com/niflaot/pixels/networking/inbound/room/floorplan/save"
	"go.uber.org/zap"
)

// NewSave creates a floor plan save packet handler.
func NewSave(handler floorplancmd.SaveHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := insave.Decode(packet)
		if err != nil {
			return err
		}
		params := domain.SaveParams{
			Heightmap: payload.Heightmap, DoorX: int(payload.DoorX), DoorY: int(payload.DoorY),
			DoorDirection: int(payload.DoorDirection), WallThickness: int(payload.WallThickness),
			FloorThickness: int(payload.FloorThickness), WallHeight: int(payload.WallHeight),
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[floorplancmd.SaveCommand]{Command: floorplancmd.SaveCommand{Handler: connection, Params: params}})
	}
}

// RegisterSave adds the floor plan save handler.
func RegisterSave(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(insave.Header, handler)
}
