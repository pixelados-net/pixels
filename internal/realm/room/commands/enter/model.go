package enter

import (
	"context"
	"errors"
	"strconv"

	roomentry "github.com/niflaot/pixels/internal/realm/room/entry"
	"github.com/niflaot/pixels/internal/realm/room/layout"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	netconn "github.com/niflaot/pixels/networking/connection"
	outentrytile "github.com/niflaot/pixels/networking/outbound/room/entrytile"
	outmodel "github.com/niflaot/pixels/networking/outbound/room/model"
	outmodelname "github.com/niflaot/pixels/networking/outbound/room/modelname"
	outrightslevel "github.com/niflaot/pixels/networking/outbound/room/rights/level"
	outrightsowner "github.com/niflaot/pixels/networking/outbound/room/rights/owner"
)

const (
	// DefaultWallHeight stores the initial room wall height.
	DefaultWallHeight int32 = 0
)

// SendModel sends the initial room model name and geometry packets.
func SendModel(ctx context.Context, connection netconn.Context, room roommodel.Room, roomLayout layout.Layout) error {
	namePacket, err := outmodelname.Encode(room.ModelName, int32(room.ID))
	if err != nil {
		return err
	}
	if err := connection.Send(ctx, namePacket); err != nil {
		return err
	}

	return SendGeometry(ctx, connection, roomLayout)
}

// entryErrorCode maps internal entry errors to protocol codes.
func entryErrorCode(err error) (int32, bool) {
	switch {
	case errors.Is(err, roomlive.ErrRoomFull):
		return ErrorRoomFull, true
	case errors.Is(err, roomentry.ErrBanned):
		return ErrorBanned, true
	case errors.Is(err, roomentry.ErrAccessDenied):
		return ErrorAccessDenied, true
	default:
		return 0, false
	}
}

// sendRights sends the current room control level to an entering player.
func (handler Handler) sendRights(ctx context.Context, connection netconn.Context, room roommodel.Room, active *roomlive.Room, playerID int64) error {
	level := outrightslevel.None
	if room.OwnerPlayerID == playerID {
		packet, err := outrightsowner.Encode()
		if err != nil {
			return err
		}
		if err := connection.Send(ctx, packet); err != nil {
			return err
		}
		level = outrightslevel.Owner
	} else if active.HasRights(playerID) {
		level = outrightslevel.Rights
	}
	packet, err := outrightslevel.Encode(level)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// SendGeometry sends entry tile and heightmap packets without retriggering Nitro's model request.
func SendGeometry(ctx context.Context, connection netconn.Context, roomLayout layout.Layout) error {
	if err := SendEntryTile(ctx, connection, roomLayout); err != nil {
		return err
	}

	modelPacket, err := outmodel.Encode(true, DefaultWallHeight, roomLayout.Heightmap)
	if err != nil {
		return err
	}

	return connection.Send(ctx, modelPacket)
}

// SendEntryTile sends the current room entry tile settings.
func SendEntryTile(ctx context.Context, connection netconn.Context, roomLayout layout.Layout) error {
	entryPacket, err := outentrytile.Encode(
		int32(roomLayout.DoorX),
		int32(roomLayout.DoorY),
		strconv.FormatFloat(float64(roomLayout.DoorZ), 'f', 1, 64),
		int32(roomLayout.DoorDirection),
	)
	if err != nil {
		return err
	}

	return connection.Send(ctx, entryPacket)
}
