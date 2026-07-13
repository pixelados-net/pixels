package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
)

// read returns one player by id with conditional ETag support.
func (handler handler) read(ctx *fiber.Ctx) error {
	id, err := playerID(ctx)
	if err != nil {
		return err
	}

	record, found, findErr := handler.players.FindByID(ctx.Context(), id)
	return handler.respondRecord(ctx, record, found, findErr)
}

// readByUsername returns one active player by exact case-insensitive username.
func (handler handler) readByUsername(ctx *fiber.Ctx) error {
	record, found, err := handler.players.FindByUsername(ctx.Context(), ctx.Params("username"))
	return handler.respondRecord(ctx, record, found, err)
}

// respondRecord writes one conditional player response.
func (handler handler) respondRecord(ctx *fiber.Ctx, record playerservice.Record, found bool, err error) error {
	if err != nil {
		return playerError(err)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "player not found")
	}
	response := playerResponse(record)
	if ctx.Get(fiber.HeaderIfNoneMatch) == etag(response) {
		return ctx.SendStatus(fiber.StatusNotModified)
	}

	setPlayerHeaders(ctx, response)
	return ctx.JSON(response)
}

// playerID parses one positive player route identifier.
func playerID(ctx *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid player id")
	}

	return id, nil
}
