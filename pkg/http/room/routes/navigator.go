package routes

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
)

// navigatorHistoryAdmin reads and removes retained navigator history.
type navigatorHistoryAdmin interface {
	// DeleteVisitHistory deletes all retained visits for one player.
	DeleteVisitHistory(context.Context, int64) (int64, error)
}

// NavigatorRoomIDsResponse contains one bounded player room identifier list.
type NavigatorRoomIDsResponse struct {
	// PlayerID identifies the requested player.
	PlayerID int64 `json:"playerId"`
	// RoomIDs stores ordered room identifiers.
	RoomIDs []int64 `json:"roomIds"`
}

// NavigatorDeleteResponse reports deleted navigator rows.
type NavigatorDeleteResponse struct {
	// PlayerID identifies the requested player.
	PlayerID int64 `json:"playerId"`
	// Deleted stores removed visit rows.
	Deleted int64 `json:"deleted"`
}

// OfficialRoomRequest attributes one administrative staff-pick mutation.
type OfficialRoomRequest struct {
	// ExpectedVersion stores the optimistic room version.
	ExpectedVersion int64 `json:"expectedVersion"`
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the administrative mutation.
	Reason string `json:"reason"`
}

// NavigatorHistoryDeleteRequest attributes one retained-history deletion.
type NavigatorHistoryDeleteRequest struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the administrative mutation.
	Reason string `json:"reason"`
}

// navigatorHistoryHandler returns recent history for one player.
func navigatorHistoryHandler(navigator navservice.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := navigatorPlayerID(ctx)
		if err != nil {
			return err
		}
		ids, err := navigator.ListRecentRoomIDs(ctx.Context(), playerID, 100)
		if err != nil {
			return err
		}
		return ctx.JSON(NavigatorRoomIDsResponse{PlayerID: playerID, RoomIDs: ids})
	}
}

// deleteNavigatorHistoryHandler removes one player's retained history.
func deleteNavigatorHistoryHandler(navigator navservice.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := navigatorPlayerID(ctx)
		if err != nil {
			return err
		}
		admin, ok := navigator.(navigatorHistoryAdmin)
		if !ok {
			return fiber.NewError(fiber.StatusNotImplemented, "navigator history administration unavailable")
		}
		var request NavigatorHistoryDeleteRequest
		if err = ctx.BodyParser(&request); err != nil || request.ActorPlayerID <= 0 || strings.TrimSpace(request.Reason) == "" {
			return fiber.NewError(fiber.StatusBadRequest, "invalid navigator history deletion request")
		}
		deleted, err := admin.DeleteVisitHistory(ctx.Context(), playerID)
		if err != nil {
			return err
		}
		return ctx.JSON(NavigatorDeleteResponse{PlayerID: playerID, Deleted: deleted})
	}
}

// navigatorFavoritesHandler returns favorite room identifiers for one player.
func navigatorFavoritesHandler(navigator navservice.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := navigatorPlayerID(ctx)
		if err != nil {
			return err
		}
		ids, err := navigator.ListFavoriteRoomIDs(ctx.Context(), playerID)
		if err != nil {
			return err
		}
		return ctx.JSON(NavigatorRoomIDsResponse{PlayerID: playerID, RoomIDs: ids})
	}
}

// officialRoomHandler sets or clears one room's official selection.
func officialRoomHandler(rooms roomservice.ConfigManager, selected bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if rooms == nil {
			return fiber.NewError(fiber.StatusNotImplemented, "official room administration unavailable")
		}
		roomID, err := strconv.ParseInt(ctx.Params("roomId"), 10, 64)
		if err != nil || roomID <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid room id")
		}
		var request OfficialRoomRequest
		if err = ctx.BodyParser(&request); err != nil || request.ExpectedVersion <= 0 || request.ActorPlayerID <= 0 || strings.TrimSpace(request.Reason) == "" {
			return fiber.NewError(fiber.StatusBadRequest, "invalid official room request")
		}
		room, err := rooms.Update(ctx.Context(), roomID, request.ExpectedVersion, roomservice.UpdateParams{StaffPicked: &selected})
		if err != nil {
			return err
		}
		return ctx.JSON(roomResponse(room))
	}
}

// navigatorPlayerID parses the shared navigator administration query.
func navigatorPlayerID(ctx *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(ctx.Query("playerId"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid player id")
	}
	return id, nil
}
