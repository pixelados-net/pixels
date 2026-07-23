package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/realm/chat/history"
	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
)

// roomHistory reads chat history for one room.
func roomHistory(service *history.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := positiveID(ctx.Params("id"), "room")
		if err != nil {
			return err
		}
		query, err := historyQuery(ctx)
		if err != nil {
			return err
		}
		query.RoomID = &roomID

		return historyResponse(ctx, service, query)
	}
}

// playerHistory reads chat history for one player and optional room.
func playerHistory(service *history.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := positiveID(ctx.Params("playerId"), "player")
		if err != nil {
			return err
		}
		query, err := historyQuery(ctx)
		if err != nil {
			return err
		}
		query.PlayerID = &playerID
		if roomValue := ctx.Query("roomId"); roomValue != "" {
			roomID, parseErr := positiveID(roomValue, "room")
			if parseErr != nil {
				return parseErr
			}
			query.RoomID = &roomID
		}

		return historyResponse(ctx, service, query)
	}
}

// historyQuery parses keyset pagination inputs.
func historyQuery(ctx *fiber.Ctx) (historymodel.Query, error) {
	query := historymodel.Query{Limit: ctx.QueryInt("limit", 50)}
	if beforeValue := ctx.Query("before"); beforeValue != "" {
		before, err := positiveID(beforeValue, "history cursor")
		if err != nil {
			return historymodel.Query{}, err
		}
		query.Before = &before
	}

	return query.Normalize(), nil
}

// historyResponse queries and writes one history page.
func historyResponse(ctx *fiber.Ctx, service *history.Service, query historymodel.Query) error {
	items, err := service.History(ctx.Context(), query)
	if err != nil {
		return err
	}

	return ctx.JSON(HistoryResponse{Total: len(items), Items: items})
}

// positiveID parses one positive route or query identity.
func positiveID(value string, label string) (int64, error) {
	identity, err := strconv.ParseInt(value, 10, 64)
	if err != nil || identity <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+label+" id")
	}

	return identity, nil
}
