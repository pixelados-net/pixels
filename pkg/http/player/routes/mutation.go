package routes

import (
	"context"

	"github.com/gofiber/fiber/v2"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// update applies one partial player identity and profile mutation.
func (handler handler) update(ctx *fiber.Ctx) error {
	id, err := playerID(ctx)
	if err != nil {
		return err
	}
	var request UpdateRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid player patch body")
	}
	if request.ClearHomeRoom && request.HomeRoomID != nil {
		return fiber.NewError(fiber.StatusBadRequest, "homeRoomId and clearHomeRoom are mutually exclusive")
	}

	record, err := handler.players.Update(ctx.Context(), id, updateParams(request))
	if err != nil {
		return playerError(err)
	}
	handler.project(record)
	response := playerResponse(record)
	setPlayerHeaders(ctx, response)

	return ctx.JSON(response)
}

// softDelete marks one player deleted and closes an active session.
func (handler handler) softDelete(ctx *fiber.Ctx) error {
	id, err := playerID(ctx)
	if err != nil {
		return err
	}
	if err := handler.players.SoftDelete(ctx.Context(), id); err != nil {
		return playerError(err)
	}
	handler.disconnectDeleted(ctx.Context(), id)

	return ctx.SendStatus(fiber.StatusNoContent)
}

// updateParams maps the HTTP patch into service input.
func updateParams(request UpdateRequest) playerservice.UpdateParams {
	params := playerservice.UpdateParams{Username: request.Username, Look: request.Look, Motto: request.Motto,
		AllowNameChange: request.AllowNameChange, BubbleStyle: request.BubbleStyle,
		BlockFriendRequests: request.BlockFriendRequests, BlockRoomInvites: request.BlockRoomInvites,
		BlockFollowing: request.BlockFollowing}
	if request.Gender != nil {
		gender := playermodel.Gender(*request.Gender)
		params.Gender = &gender
	}
	if request.ClearHomeRoom {
		var homeRoomID *int64
		params.HomeRoomID = &homeRoomID
	} else if request.HomeRoomID != nil {
		params.HomeRoomID = &request.HomeRoomID
	}

	return params
}

// project replaces one online player's durable snapshot.
func (handler handler) project(record playerservice.Record) {
	if handler.live == nil {
		return
	}
	player, found := handler.live.Find(record.Player.ID)
	if found {
		_ = player.ReplaceSnapshot(playerlive.SnapshotFromRecord(record))
	}
}

// disconnectDeleted closes one deleted player's active connection when present.
func (handler handler) disconnectDeleted(ctx context.Context, playerID int64) {
	if handler.live == nil || handler.connections == nil {
		return
	}
	player, found := handler.live.Find(playerID)
	if !found {
		return
	}
	peer := player.Peer()
	_ = handler.connections.Disconnect(ctx, peer.ConnectionKind(), peer.ConnectionID(), netconn.Reason{
		Code: netconn.DisconnectPolicyViolation, Message: "player account deleted",
	})
}
