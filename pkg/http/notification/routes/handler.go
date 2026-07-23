package routes

import (
	"github.com/gofiber/fiber/v2"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/i18n"
)

const (
	// KindBubble sends a BUBBLE_ALERT packet.
	KindBubble = "bubble"

	// KindAlert sends a GENERIC_ALERT packet.
	KindAlert = "alert"

	// defaultBubbleKey stores the generic admin bubble type.
	defaultBubbleKey = "admin.notification"
)

// notifyHandler sends one localized notification to a live player.
func notifyHandler(players *playerlive.Registry, connections *netconn.Registry, translations i18n.Translator) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		request, err := parseNotifyRequest(ctx)
		if err != nil {
			return err
		}

		connection, err := playerConnection(players, connections, request.PlayerID)
		if err != nil {
			return err
		}

		packet, err := notificationPacket(request, translations)
		if err != nil {
			return err
		}
		if err := connection.Send(ctx.Context(), packet); err != nil {
			return err
		}

		return ctx.JSON(NotifyResponse{PlayerID: request.PlayerID, Kind: request.kind(), Key: request.Key, Sent: true})
	}
}

// parseNotifyRequest parses and validates notification input.
func parseNotifyRequest(ctx *fiber.Ctx) (NotifyRequest, error) {
	var request NotifyRequest
	if err := ctx.BodyParser(&request); err != nil {
		return NotifyRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid notification request body")
	}
	if request.PlayerID <= 0 {
		return NotifyRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid player id")
	}
	if request.Key == "" {
		return NotifyRequest{}, fiber.NewError(fiber.StatusBadRequest, "notification key is required")
	}

	return request, nil
}

// playerConnection resolves a live player's connection.
func playerConnection(players *playerlive.Registry, connections *netconn.Registry, playerID int64) (netconn.Connection, error) {
	player, found := players.Find(playerID)
	if !found {
		return nil, fiber.NewError(fiber.StatusNotFound, "player not online")
	}

	peer := player.Peer()
	connection, found := connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if !found {
		return nil, fiber.NewError(fiber.StatusNotFound, "player connection not found")
	}

	return connection, nil
}
