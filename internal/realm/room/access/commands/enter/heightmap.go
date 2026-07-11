package enter

import (
	"context"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	netconn "github.com/niflaot/pixels/networking/connection"
	outheightmap "github.com/niflaot/pixels/networking/outbound/room/heightmap"
)

// sendHeightMap sends the current room surface height map to one connection.
func (handler Handler) sendHeightMap(ctx context.Context, connection netconn.Context, active *roomlive.Room) error {
	if active == nil {
		return nil
	}

	width, _, tiles := active.SurfaceHeights()
	if len(tiles) == 0 {
		return nil
	}

	packet, err := outheightmap.Encode(int(width), projection.HeightMapTiles(tiles))
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
