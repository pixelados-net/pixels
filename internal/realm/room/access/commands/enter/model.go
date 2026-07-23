package enter

import (
	"context"
	"errors"
	"strconv"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	outentered "github.com/niflaot/pixels/networking/outbound/room/entered"
	outentryinfo "github.com/niflaot/pixels/networking/outbound/room/entryinfo"
	outentrytile "github.com/niflaot/pixels/networking/outbound/room/entrytile"
	outmodel "github.com/niflaot/pixels/networking/outbound/room/model"
	outmodelname "github.com/niflaot/pixels/networking/outbound/room/modelname"
	outrightslevel "github.com/niflaot/pixels/networking/outbound/room/rights/level"
	outrightsowner "github.com/niflaot/pixels/networking/outbound/room/rights/owner"
	outscore "github.com/niflaot/pixels/networking/outbound/room/score"
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

// sendEntered sends the initial room entry packets.
func (handler Handler) sendEntered(ctx context.Context, connection netconn.Context, room roommodel.Room, roomLayout layout.Layout, active *roomlive.Room, playerID int64) error {
	packet, err := outentered.Encode()
	if err != nil {
		return err
	}
	if err := connection.Send(ctx, packet); err != nil {
		return err
	}
	if err := SendModel(ctx, connection, room, roomLayout); err != nil {
		return err
	}
	if err := handler.sendFloorItems(ctx, connection, room, active); err != nil {
		return err
	}
	if err := sendAppearance(ctx, connection, room); err != nil {
		return err
	}
	if err := handler.sendHeightMap(ctx, connection, active); err != nil {
		return err
	}
	if err := handler.sendRoomState(ctx, connection, active, 0); err != nil {
		return err
	}
	if err := sendEntryInfo(ctx, connection, room, playerID); err != nil {
		return err
	}
	if err := handler.sendRights(ctx, connection, room, active, playerID); err != nil {
		return err
	}

	return handler.sendVoteState(ctx, connection, room.ID, playerID)
}

// sendEntryInfo identifies the entered room and its persistent owner to Nitro.
func sendEntryInfo(ctx context.Context, connection netconn.Context, room roommodel.Room, playerID int64) error {
	packet, err := outentryinfo.Encode(int32(room.ID), room.OwnerPlayerID == playerID)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// sendVoteState sends current room score and vote eligibility.
func (handler Handler) sendVoteState(ctx context.Context, connection netconn.Context, roomID int64, playerID int64) error {
	if handler.Votes == nil {
		return nil
	}
	state, err := handler.Votes.State(ctx, roomID, playerID)
	if err != nil {
		return err
	}
	packet, err := outscore.Encode(int32(state.Score), state.CanVote)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
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
	} else if active.CanManageFurniture(playerID) {
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

	wallHeight := DefaultWallHeight
	if roomLayout.RoomID > 0 {
		wallHeight = int32(roomLayout.WallHeight)
	}
	modelPacket, err := outmodel.Encode(true, wallHeight, roomLayout.Heightmap)
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
		int32(roomLayout.DoorDirection),
	)
	if err != nil {
		return err
	}

	return connection.Send(ctx, entryPacket)
}

// roomSnapshot maps persistent rooms to runtime snapshots.
func roomSnapshot(room roommodel.Room) roomlive.Snapshot {
	return roomlive.Snapshot{
		ID: room.ID, OwnerPlayerID: room.OwnerPlayerID,
		CategoryID: room.CategoryID, MaxUsers: room.MaxUsers, TradeMode: int16(room.TradeMode),
		RollerSpeed:  room.RollerSpeed,
		ChatDistance: room.ChatDistance, ChatProtection: room.ChatProtection,
		AllowPets: room.AllowPets, AllowPetsEat: room.AllowPetsEat,
	}
}
