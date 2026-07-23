package routes

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/realm/chat/bubble"
)

// listBubbles lists configured bubble thresholds.
func listBubbles(service *bubble.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		items, err := service.List(ctx.Context())
		if err != nil {
			return err
		}

		return ctx.JSON(BubbleListResponse{Total: len(items), Items: items})
	}
}

// setBubble creates or replaces one bubble threshold.
func setBubble(service *bubble.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		bubbleID, err := bubbleIDParam(ctx)
		if err != nil {
			return err
		}
		var request BubbleRequest
		if err = ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid chat bubble request")
		}
		if err = service.SetUnlock(ctx.Context(), bubbleID, request.MinWeight); err != nil {
			return mapBubbleError(err)
		}

		return ctx.JSON(MutationResponse{Updated: true})
	}
}

// deleteBubble removes one bubble threshold.
func deleteBubble(service *bubble.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		bubbleID, err := bubbleIDParam(ctx)
		if err != nil {
			return err
		}
		if err = service.DeleteUnlock(ctx.Context(), bubbleID); err != nil {
			return mapBubbleError(err)
		}

		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

// bubbleIDParam parses a protocol bubble id.
func bubbleIDParam(ctx *fiber.Ctx) (int32, error) {
	value, err := strconv.ParseInt(ctx.Params("bubbleId"), 10, 32)
	if err != nil || value < 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid chat bubble id")
	}

	return int32(value), nil
}

// mapBubbleError maps expected bubble validation failures.
func mapBubbleError(err error) error {
	if errors.Is(err, bubble.ErrInvalidBubble) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return err
}
