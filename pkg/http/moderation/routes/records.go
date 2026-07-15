package routes

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
)

// issues lists bounded moderation issues for external tooling.
func (dependencies Dependencies) issues(ctx *fiber.Ctx) error {
	state := strings.TrimSpace(ctx.Query("state"))
	if state != "" && state != "open" && state != "picked" && state != "resolved" && state != "deleted" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid issue state")
	}
	limit, err := parseLimit(ctx, 100, 500)
	if err != nil {
		return err
	}
	items, err := dependencies.Moderation.Store().Issues(ctx.Context(), state, limit)
	if err != nil {
		return routeError(err)
	}
	return ctx.JSON(fiber.Map{"items": items, "count": len(items)})
}

// topics lists all call-for-help topics, including disabled records.
func (dependencies Dependencies) topics(ctx *fiber.Ctx) error {
	items, err := dependencies.Moderation.Store().Topics(ctx.Context(), true)
	if err != nil {
		return routeError(err)
	}
	return ctx.JSON(fiber.Map{"items": items, "count": len(items)})
}

// createTopic creates and publishes one call-for-help topic.
func (dependencies Dependencies) createTopic(ctx *fiber.Ctx) error {
	var body TopicRequest
	if err := request(ctx, &body); err != nil {
		return err
	}
	value := moderationrecord.Topic{Category: body.Category, NameKey: body.NameKey, Action: body.Action, AutoReplyKey: body.AutoReplyKey, DefaultSanctionLadder: body.DefaultSanctionLadder, Order: body.Order, Enabled: body.Enabled}
	normalizeTopic(&value)
	if err := validateTopic(value); err != nil {
		return err
	}
	value, err := dependencies.Moderation.Store().CreateTopic(ctx.Context(), value)
	if err != nil {
		return routeError(err)
	}
	if err = dependencies.Moderation.RefreshTopics(ctx.Context()); err != nil {
		return routeError(err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(value)
}

// patchTopic partially updates and republishes one call-for-help topic.
func (dependencies Dependencies) patchTopic(ctx *fiber.Ctx) error {
	id, err := pathID(ctx, "id")
	if err != nil {
		return err
	}
	value, found, err := dependencies.Moderation.Store().Topic(ctx.Context(), id)
	if err != nil {
		return routeError(err)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "moderation topic not found")
	}
	var body TopicPatchRequest
	if err = request(ctx, &body); err != nil {
		return err
	}
	mergeTopic(&value, body)
	normalizeTopic(&value)
	if err = validateTopic(value); err != nil {
		return err
	}
	value, found, err = dependencies.Moderation.Store().UpdateTopic(ctx.Context(), value)
	if err != nil {
		return routeError(err)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "moderation topic not found")
	}
	if err = dependencies.Moderation.RefreshTopics(ctx.Context()); err != nil {
		return routeError(err)
	}
	return ctx.JSON(value)
}

// validateTopic validates persistence and client behavior constraints.
func validateTopic(value moderationrecord.Topic) error {
	if value.Category == "" || value.NameKey == "" || value.Action != "queue" && value.Action != "auto_reply" && value.Action != "ignore" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid moderation topic")
	}
	if value.Action == "auto_reply" && (value.AutoReplyKey == nil || strings.TrimSpace(*value.AutoReplyKey) == "") {
		return fiber.NewError(fiber.StatusBadRequest, "autoReplyKey is required for auto_reply topics")
	}
	return nil
}

// normalizeTopic stores canonical topic fields.
func normalizeTopic(value *moderationrecord.Topic) {
	value.Category = strings.TrimSpace(value.Category)
	value.NameKey = strings.TrimSpace(value.NameKey)
	value.Action = strings.TrimSpace(value.Action)
	if value.AutoReplyKey != nil {
		normalized := strings.TrimSpace(*value.AutoReplyKey)
		value.AutoReplyKey = &normalized
	}
}

// mergeTopic applies optional topic fields.
func mergeTopic(value *moderationrecord.Topic, body TopicPatchRequest) {
	if body.Category != nil {
		value.Category = *body.Category
	}
	if body.NameKey != nil {
		value.NameKey = *body.NameKey
	}
	if body.Action != nil {
		value.Action = *body.Action
	}
	if body.AutoReplyKey != nil {
		value.AutoReplyKey = *body.AutoReplyKey
	}
	if body.DefaultSanctionLadder != nil {
		value.DefaultSanctionLadder = *body.DefaultSanctionLadder
	}
	if body.Order != nil {
		value.Order = *body.Order
	}
	if body.Enabled != nil {
		value.Enabled = *body.Enabled
	}
}
