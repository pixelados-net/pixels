package routes

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	"go.uber.org/fx"
)

// Dependencies contains global moderation administration behavior.
type Dependencies struct {
	fx.In

	// Moderation manages issues, topics, and presets.
	Moderation *moderationcore.Service
	// Sanctions manages global punishments and escalation.
	Sanctions *sanctioncore.Service
}

// Register mounts protected global moderation routes.
func Register(app *fiber.App, dependencies Dependencies) {
	if dependencies.Moderation == nil || dependencies.Sanctions == nil {
		return
	}
	app.Get("/api/admin/players/:playerId/punishments", dependencies.punishments)
	app.Post("/api/admin/players/:playerId/punishments", dependencies.applyPunishment)
	app.Delete("/api/admin/punishments/:id", dependencies.revokePunishment)
	app.Get("/api/admin/moderation/issues", dependencies.issues)
	app.Get("/api/admin/moderation/cfh-topics", dependencies.topics)
	app.Post("/api/admin/moderation/cfh-topics", dependencies.createTopic)
	app.Patch("/api/admin/moderation/cfh-topics/:id", dependencies.patchTopic)
	app.Get("/api/admin/moderation/presets", dependencies.presets)
	app.Post("/api/admin/moderation/presets", dependencies.createPreset)
	app.Patch("/api/admin/moderation/presets/:id", dependencies.patchPreset)
	app.Get("/api/admin/moderation/sanction-ladder", dependencies.ladder)
	app.Put("/api/admin/moderation/sanction-ladder", dependencies.replaceLadder)
}

// pathID parses one positive route identifier.
func pathID(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid path identifier")
	}
	return value, nil
}

// parseLimit parses one bounded query limit.
func parseLimit(ctx *fiber.Ctx, fallback int32, maximum int32) (int32, error) {
	value := strings.TrimSpace(ctx.Query("limit"))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil || parsed <= 0 || parsed > int64(maximum) {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid limit")
	}
	return int32(parsed), nil
}

// request parses one required JSON request body.
func request(ctx *fiber.Ctx, destination any) error {
	if err := ctx.BodyParser(destination); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	return nil
}

// routeError maps domain errors to meaningful HTTP status codes.
func routeError(err error) error {
	switch {
	case errors.Is(err, sanctioncore.ErrInvalidRequest), errors.Is(err, moderationcore.ErrInvalid):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, sanctioncore.ErrUnauthorized), errors.Is(err, sanctioncore.ErrImmune), errors.Is(err, moderationcore.ErrUnauthorized):
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	case errors.Is(err, sanctioncore.ErrNotFound), errors.Is(err, moderationcore.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, moderationcore.ErrPickFailed):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	default:
		return err
	}
}
