package routes

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
)

// presets lists every moderator response preset.
func (dependencies Dependencies) presets(ctx *fiber.Ctx) error {
	items, err := dependencies.Moderation.Store().Presets(ctx.Context(), true)
	if err != nil {
		return routeError(err)
	}
	return ctx.JSON(fiber.Map{"items": items, "count": len(items)})
}

// createPreset creates one localized moderator response preset.
func (dependencies Dependencies) createPreset(ctx *fiber.Ctx) error {
	var body PresetRequest
	if err := request(ctx, &body); err != nil {
		return err
	}
	value := moderationrecord.Preset{Category: strings.TrimSpace(body.Category), MessageKey: strings.TrimSpace(body.MessageKey), Enabled: body.Enabled, Order: body.Order}
	if value.Category == "" || value.MessageKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category and messageKey are required")
	}
	value, err := dependencies.Moderation.Store().CreatePreset(ctx.Context(), value)
	if err != nil {
		return routeError(err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(value)
}

// patchPreset partially updates one moderator response preset.
func (dependencies Dependencies) patchPreset(ctx *fiber.Ctx) error {
	id, err := pathID(ctx, "id")
	if err != nil {
		return err
	}
	items, err := dependencies.Moderation.Store().Presets(ctx.Context(), true)
	if err != nil {
		return routeError(err)
	}
	var value moderationrecord.Preset
	found := false
	for _, candidate := range items {
		if candidate.ID == id {
			value, found = candidate, true
			break
		}
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "moderation preset not found")
	}
	var body PresetPatchRequest
	if err = request(ctx, &body); err != nil {
		return err
	}
	if body.Category != nil {
		value.Category = strings.TrimSpace(*body.Category)
	}
	if body.MessageKey != nil {
		value.MessageKey = strings.TrimSpace(*body.MessageKey)
	}
	if body.Enabled != nil {
		value.Enabled = *body.Enabled
	}
	if body.Order != nil {
		value.Order = *body.Order
	}
	if value.Category == "" || value.MessageKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category and messageKey are required")
	}
	value, found, err = dependencies.Moderation.Store().UpdatePreset(ctx.Context(), value)
	if err != nil {
		return routeError(err)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "moderation preset not found")
	}
	return ctx.JSON(value)
}

// ladder lists ordered global sanction escalation policy.
func (dependencies Dependencies) ladder(ctx *fiber.Ctx) error {
	items, err := dependencies.Sanctions.Store().Ladder(ctx.Context())
	if err != nil {
		return routeError(err)
	}
	return ctx.JSON(fiber.Map{"items": items, "count": len(items)})
}

// replaceLadder validates and atomically replaces escalation policy.
func (dependencies Dependencies) replaceLadder(ctx *fiber.Ctx) error {
	var body LadderRequest
	if err := request(ctx, &body); err != nil {
		return err
	}
	if len(body.Items) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "sanction ladder cannot be empty")
	}
	entries := make([]sanctionrecord.LadderEntry, len(body.Items))
	for index, item := range body.Items {
		kind := sanctionrecord.Kind(item.Kind)
		if item.Level != int32(index+1) || kind != sanctionrecord.KindWarn && kind != sanctionrecord.KindMute && kind != sanctionrecord.KindBan || item.DurationHours < 0 || item.DurationHours > sanctionrecord.MaxDurationHours || item.ProbationDays <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid contiguous sanction ladder")
		}
		entries[index] = sanctionrecord.LadderEntry{Level: item.Level, Kind: kind, DurationHours: item.DurationHours, ProbationDays: item.ProbationDays}
	}
	if err := dependencies.Sanctions.Store().ReplaceLadder(ctx.Context(), entries); err != nil {
		return routeError(err)
	}
	return ctx.JSON(fiber.Map{"items": entries, "count": len(entries)})
}
