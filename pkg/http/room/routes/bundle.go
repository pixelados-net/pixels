package routes

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
)

// bundleTemplatesHandler lists marked room bundle templates.
func bundleTemplatesHandler(bundles roombundle.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		rooms, err := bundles.Templates(ctx.Context())
		if err != nil {
			return err
		}
		items := roomResponses(rooms)
		return ctx.JSON(ListResponse{Total: len(items), Items: items})
	}
}

// markBundleTemplateHandler marks one room as a bundle template.
func markBundleTemplateHandler(bundles roombundle.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := bundleRoomID(ctx)
		if err != nil {
			return err
		}
		room, err := bundles.Mark(ctx.Context(), roomID)
		if err != nil {
			return bundleError(err)
		}
		return ctx.JSON(roomResponse(room))
	}
}

// unmarkBundleTemplateHandler removes bundle template status.
func unmarkBundleTemplateHandler(bundles roombundle.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := bundleRoomID(ctx)
		if err != nil {
			return err
		}
		room, err := bundles.Unmark(ctx.Context(), roomID)
		if err != nil {
			return bundleError(err)
		}
		return ctx.JSON(roomResponse(room))
	}
}

// bundleRoomID parses a positive room route identifier.
func bundleRoomID(ctx *fiber.Ctx) (int64, error) {
	roomID, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || roomID <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid room id")
	}
	return roomID, nil
}

// bundleError maps room bundle administration failures.
func bundleError(err error) error {
	switch {
	case errors.Is(err, roombundle.ErrRoomNotFound), errors.Is(err, roombundle.ErrInvalidTemplate):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, roombundle.ErrTemplateReferenced):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	default:
		return err
	}
}
