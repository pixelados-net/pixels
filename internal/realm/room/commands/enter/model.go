package enter

import (
	"context"
	"errors"
	"strconv"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	roomentry "github.com/niflaot/pixels/internal/realm/room/entry"
	"github.com/niflaot/pixels/internal/realm/room/layout"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
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

// ControlPolicy maps global room capabilities to Nitro controller state.
type ControlPolicy struct {
	// Permissions resolves the entering player's global capabilities.
	Permissions permissionservice.Checker
	// RightsAnyGrant allows granting build rights in any room.
	RightsAnyGrant permission.Node
	// RightsAnyRevoke allows revoking build rights in any room.
	RightsAnyRevoke permission.Node
	// ModerationAnyKick allows kicking in any room.
	ModerationAnyKick permission.Node
	// ModerationAnyMute allows muting in any room.
	ModerationAnyMute permission.Node
	// ModerationAnyBan allows banning in any room.
	ModerationAnyBan permission.Node
}

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
	staffRights, staffModeration, err := handler.Control.Resolve(ctx, playerID)
	if err != nil {
		return err
	}

	level := outrightslevel.None
	isOwner := room.OwnerPlayerID == playerID
	if isOwner || staffRights {
		packet, err := outrightsowner.Encode()
		if err != nil {
			return err
		}
		if err := connection.Send(ctx, packet); err != nil {
			return err
		}
	}
	if staffModeration {
		level = outrightslevel.Moderator
	} else if isOwner {
		level = outrightslevel.Owner
	} else if active.HasRights(playerID) {
		level = outrightslevel.Rights
	}
	active.SetUnitStatus(playerID, worldunit.StatusFlatControl, strconv.Itoa(int(level)))
	packet, err := outrightslevel.Encode(level)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// Resolve reports whether a player has global rights or moderation control.
func (policy ControlPolicy) Resolve(ctx context.Context, playerID int64) (bool, bool, error) {
	rights, err := policy.allowsAny(ctx, playerID, policy.RightsAnyGrant, policy.RightsAnyRevoke)
	if err != nil {
		return false, false, err
	}
	moderation, err := policy.allowsAny(ctx, playerID, policy.ModerationAnyKick, policy.ModerationAnyMute, policy.ModerationAnyBan)
	if err != nil {
		return false, false, err
	}

	return rights, moderation, nil
}

// allowsAny reports whether a player holds at least one supplied capability.
func (policy ControlPolicy) allowsAny(ctx context.Context, playerID int64, nodes ...permission.Node) (bool, error) {
	if policy.Permissions == nil {
		return false, nil
	}
	for _, node := range nodes {
		if node == "" {
			continue
		}
		allowed, err := policy.Permissions.HasPermission(ctx, playerID, node)
		if err != nil || allowed {
			return allowed, err
		}
	}

	return false, nil
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
