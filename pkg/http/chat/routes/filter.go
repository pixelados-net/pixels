package routes

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
)

// listFilters lists the global chat dictionary.
func listFilters(service *chatfilter.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		words := service.List()
		return ctx.JSON(FilterListResponse{Total: len(words), Items: words})
	}
}

// addFilter adds one global chat filter word.
func addFilter(service *chatfilter.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request FilterRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid global chat filter request")
		}
		if err := service.Add(ctx.Context(), request.Word); err != nil {
			return mapFilterError(err)
		}

		return ctx.Status(fiber.StatusCreated).JSON(MutationResponse{Updated: true})
	}
}

// removeFilter removes one global chat filter word.
func removeFilter(service *chatfilter.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if err := service.Remove(ctx.Context(), ctx.Params("word")); err != nil {
			return mapFilterError(err)
		}

		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

// mapFilterError maps expected filter validation failures.
func mapFilterError(err error) error {
	if errors.Is(err, chatfilter.ErrInvalidWord) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return err
}
