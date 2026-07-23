package routes

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
)

// teleportHandler forwards one live player to a target room.
func teleportHandler(rooms roomservice.Manager, players *playerlive.Registry, connections *netconn.Registry, entry *roomentry.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, request, err := teleportInput(ctx)
		if err != nil {
			return err
		}
		target, found, err := rooms.FindByID(ctx.Context(), request.TargetRoomID)
		if err != nil {
			return err
		}
		if !found {
			return fiber.NewError(fiber.StatusNotFound, "target room not found")
		}
		player, found := players.Find(playerID)
		if !found {
			return fiber.NewError(fiber.StatusNotFound, "live player not found")
		}
		peer := player.Peer()
		connection, found := connections.Get(peer.ConnectionKind(), peer.ConnectionID())
		if !found {
			return fiber.NewError(fiber.StatusNotFound, "player connection not found")
		}
		if request.Bypass && target.DoorMode != roommodel.DoorModeOpen && (entry == nil || !entry.GrantTrusted(playerID, request.TargetRoomID)) {
			return fiber.NewError(fiber.StatusInternalServerError, "room entry bypass could not be granted")
		}
		packet, err := outforward.Encode(int32(request.TargetRoomID))
		if err != nil {
			return err
		}
		if err := connection.Send(ctx.Context(), packet); err != nil {
			return err
		}

		return ctx.JSON(ActionResponse{Matched: 1, Sent: 1})
	}
}

// teleportInput parses one player teleport request.
func teleportInput(ctx *fiber.Ctx) (int64, TeleportRequest, error) {
	playerID, err := strconv.ParseInt(ctx.Params("playerId"), 10, 64)
	if err != nil || playerID <= 0 {
		return 0, TeleportRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid player id")
	}
	var request TeleportRequest
	if err := ctx.BodyParser(&request); err != nil {
		return 0, TeleportRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid room teleport request body")
	}
	if request.TargetRoomID <= 0 || request.TargetRoomID > math.MaxInt32 {
		return 0, TeleportRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid target room id")
	}

	return playerID, request, nil
}
