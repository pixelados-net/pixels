package routes

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

// listServeItems returns every bartender keyword mapping.
func (dependencies Dependencies) listServeItems(ctx *fiber.Ctx) error {
	items, err := dependencies.Bots.ListServeItems(ctx.Context())
	if err != nil {
		return err
	}
	result := make([]ServeItemResponse, len(items))
	for index, item := range items {
		result[index] = serveItemResponse(item)
	}
	return ctx.JSON(ServeItemListResponse{Items: result, Count: len(result)})
}

// createServeItem creates one bartender keyword mapping.
func (dependencies Dependencies) createServeItem(ctx *fiber.Ctx) error {
	request := ServeItemRequest{}
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid serve item request body")
	}
	item, err := dependencies.Bots.CreateServeItem(ctx.Context(), request.Keyword, request.DefinitionID)
	if err != nil {
		return serveItemError(err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(serveItemResponse(item))
}

// updateServeItem replaces one bartender keyword mapping.
func (dependencies Dependencies) updateServeItem(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return err
	}
	request := ServeItemRequest{}
	if err = ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid serve item request body")
	}
	item, found, err := dependencies.Bots.UpdateServeItem(ctx.Context(), id, request.Keyword, request.DefinitionID)
	if err != nil {
		return serveItemError(err)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "serve item not found")
	}
	return ctx.JSON(serveItemResponse(item))
}

// deleteServeItem removes one bartender keyword mapping.
func (dependencies Dependencies) deleteServeItem(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return err
	}
	deleted, err := dependencies.Bots.DeleteServeItem(ctx.Context(), id)
	if err != nil {
		return serveItemError(err)
	}
	if !deleted {
		return fiber.NewError(fiber.StatusNotFound, "serve item not found")
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// positiveID parses one positive path identifier.
func positiveID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid path identifier")
	}
	return id, nil
}

// serveItemResponse maps one domain record.
func serveItemResponse(item botrecord.ServeItem) ServeItemResponse {
	return ServeItemResponse{ID: item.ID, Keyword: item.Keyword, DefinitionID: item.DefinitionID}
}

// serveItemError maps validation and conflict failures.
func serveItemError(err error) error {
	if errors.Is(err, botrecord.ErrInvalidSkill) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if errors.Is(err, botrecord.ErrServeKeywordExists) {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}
	return err
}
