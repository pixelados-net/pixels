package routes

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// auditContext attaches validated administrative attribution to a mutation.
func auditContext(ctx *fiber.Ctx, request AuditRequest) context.Context {
	return grouprecord.WithAudit(ctx.Context(), request.ActorPlayerID, strings.TrimSpace(request.Reason))
}

// positiveParam parses one positive path identifier.
func positiveParam(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid group administration identifier")
	}
	return value, nil
}

// boundedInt parses one bounded query integer.
func boundedInt(value string, fallback int, maximum int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 || parsed > maximum {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid pagination value")
	}
	return parsed, nil
}

// optionalPositive parses one optional positive query identifier.
func optionalPositive(value string) (int64, error) {
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid group administration filter")
	}
	return parsed, nil
}

// optionalBool parses one optional boolean query filter.
func optionalBool(value string) (*bool, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid group administration filter")
	}
	return &parsed, nil
}

// body parses one required JSON request.
func body(ctx *fiber.Ctx, target any) error {
	if err := ctx.BodyParser(target); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid group administration request body")
	}
	return nil
}

// requireReason validates durable mutation attribution.
func requireReason(request AuditRequest) error {
	if request.ActorPlayerID <= 0 || len(strings.TrimSpace(request.Reason)) < 3 || len(request.Reason) > 500 {
		return fiber.NewError(fiber.StatusBadRequest, "actorPlayerId and reason are required")
	}
	return nil
}

// parts maps transport badge data to normalized ordinals.
func parts(values []BadgePartRequest) []grouprecord.BadgePart {
	result := make([]grouprecord.BadgePart, len(values))
	for index, value := range values {
		result[index] = grouprecord.BadgePart{Ordinal: int16(index), Kind: value.Kind, ElementID: value.ElementID, ColorID: value.ColorID, Position: value.Position}
	}
	return result
}

// groupError maps expected domain failures to meaningful HTTP errors.
func groupError(err error) error {
	switch {
	case errors.Is(err, grouprecord.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, grouprecord.ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	case errors.Is(err, grouprecord.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, grouprecord.ErrLimit):
		return fiber.NewError(fiber.StatusTooManyRequests, err.Error())
	case errors.Is(err, grouprecord.ErrInvalid), errors.Is(err, grouprecord.ErrClosed):
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	default:
		return err
	}
}
