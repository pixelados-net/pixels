package enter

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/layout"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	netconn "github.com/niflaot/pixels/networking/connection"
	outmodel "github.com/niflaot/pixels/networking/outbound/room/model"
	outmodelname "github.com/niflaot/pixels/networking/outbound/room/modelname"
)

const (
	// DefaultWallHeight stores the initial room wall height.
	DefaultWallHeight int32 = 0
)

// SendModel sends room model name and heightmap packets.
func SendModel(ctx context.Context, connection netconn.Context, room roommodel.Room, roomLayout layout.Layout) error {
	namePacket, err := outmodelname.Encode(room.ModelName, int32(room.ID))
	if err != nil {
		return err
	}
	if err := connection.Send(ctx, namePacket); err != nil {
		return err
	}

	modelPacket, err := outmodel.Encode(false, DefaultWallHeight, roomLayout.Heightmap)
	if err != nil {
		return err
	}

	return connection.Send(ctx, modelPacket)
}
