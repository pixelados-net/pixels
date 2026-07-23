package routes

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
)

// punishments lists one player's complete punishment history.
func (dependencies Dependencies) punishments(ctx *fiber.Ctx) error {
	playerID, err := pathID(ctx, "playerId")
	if err != nil {
		return err
	}
	limit, err := parseLimit(ctx, 100, 500)
	if err != nil {
		return err
	}
	values, err := dependencies.Sanctions.History(ctx.Context(), playerID, limit)
	if err != nil {
		return routeError(err)
	}
	items := make([]PunishmentResponse, len(values))
	for index, value := range values {
		items[index] = punishmentResponse(value)
	}
	return ctx.JSON(fiber.Map{"items": items, "count": len(items)})
}

// applyPunishment applies one punishment through the shared engine.
func (dependencies Dependencies) applyPunishment(ctx *fiber.Ctx) error {
	playerID, err := pathID(ctx, "playerId")
	if err != nil {
		return err
	}
	var body ApplyPunishmentRequest
	if err = request(ctx, &body); err != nil {
		return err
	}
	if body.IssuerPlayerID <= 0 || strings.TrimSpace(body.Reason) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issuerPlayerId and reason are required")
	}
	var expiresAt *time.Time
	if body.DurationHours != nil {
		if *body.DurationHours <= 0 || *body.DurationHours > sanctionrecord.MaxDurationHours {
			return fiber.NewError(fiber.StatusBadRequest, "durationHours must be between 1 and 87600")
		}
		value := time.Now().Add(time.Duration(*body.DurationHours) * time.Hour)
		expiresAt = &value
	}
	issuerID := body.IssuerPlayerID
	value, err := dependencies.Sanctions.Apply(ctx.Context(), sanctionrecord.ApplyParams{ReceiverPlayerID: playerID, IssuerPlayerID: &issuerID, IssuerKind: "player", Kind: sanctionrecord.Kind(body.Kind), Reason: dependencies.Moderation.Sanitize(body.Reason), CFHTopicID: body.CFHTopicID, IssueID: body.IssueID, Source: "admin_http", ExpiresAt: expiresAt})
	if err != nil {
		return routeError(err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(punishmentResponse(value))
}

// revokePunishment revokes one active punishment through the shared engine.
func (dependencies Dependencies) revokePunishment(ctx *fiber.Ctx) error {
	id, err := pathID(ctx, "id")
	if err != nil {
		return err
	}
	var body RevokePunishmentRequest
	if err = request(ctx, &body); err != nil {
		return err
	}
	value, err := dependencies.Sanctions.Revoke(ctx.Context(), id, body.RevokedByPlayerID)
	if err != nil {
		return routeError(err)
	}
	return ctx.JSON(punishmentResponse(value))
}

// punishmentResponse maps one domain record to stable JSON.
func punishmentResponse(value sanctionrecord.Punishment) PunishmentResponse {
	return PunishmentResponse{ID: value.ID, ReceiverPlayerID: value.ReceiverPlayerID, IssuerPlayerID: value.IssuerPlayerID, IssuerKind: value.IssuerKind, Kind: string(value.Kind), Reason: value.Reason, CFHTopicID: value.CFHTopicID, IssueID: value.IssueID, Source: value.Source, IssuedAt: value.IssuedAt, ExpiresAt: value.ExpiresAt, RevokedAt: value.RevokedAt, RevokedByPlayerID: value.RevokedByPlayerID, Active: value.ActiveAt(time.Now())}
}
