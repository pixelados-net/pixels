package routes

import (
	"errors"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
)

// EffectRequest contains one administrative effect grant.
type EffectRequest struct {
	// EffectID identifies the Nitro effect.
	EffectID int32 `json:"effectId"`
	// DurationSeconds stores one charge duration; zero means permanent.
	DurationSeconds int32 `json:"durationSeconds"`
}

// EffectResponse contains the resulting effect stack.
type EffectResponse struct {
	// PlayerID identifies the effect owner.
	PlayerID int64 `json:"playerId"`
	// EffectID identifies the Nitro effect.
	EffectID int32 `json:"effectId"`
	// DurationSeconds stores one charge duration.
	DurationSeconds int32 `json:"durationSeconds"`
	// RemainingCharges stores the stack count.
	RemainingCharges int32 `json:"remainingCharges"`
}

// grantEffect grants one effect charge to a player.
func (handler handler) grantEffect(ctx *fiber.Ctx) error {
	playerID, err := positivePathID(ctx, "playerId")
	if err != nil {
		return err
	}
	var request EffectRequest
	if err = ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid player effect request body")
	}
	if handler.effects == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "player effects are unavailable")
	}
	effect, err := handler.effects.Grant(ctx.Context(), playerID, request.EffectID, request.DurationSeconds, playereffect.SourceAdmin)
	if err != nil {
		return effectError(err)
	}
	return ctx.JSON(EffectResponse{PlayerID: effect.PlayerID, EffectID: effect.ID, DurationSeconds: effect.DurationSeconds, RemainingCharges: effect.RemainingCharges})
}

// revokeEffect removes one player's durable effect stack.
func (handler handler) revokeEffect(ctx *fiber.Ctx) error {
	playerID, err := positivePathID(ctx, "playerId")
	if err != nil {
		return err
	}
	effectID, err := positivePathID(ctx, "effectId")
	if err != nil || effectID > math.MaxInt32 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid effect id")
	}
	if handler.effects == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "player effects are unavailable")
	}
	if err = handler.effects.Revoke(ctx.Context(), playerID, int32(effectID)); err != nil {
		return effectError(err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// positivePathID parses one positive route identifier.
func positivePathID(ctx *fiber.Ctx, name string) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return id, nil
}

// effectError maps domain effect failures into meaningful HTTP errors.
func effectError(err error) error {
	if errors.Is(err, playereffect.ErrInvalidEffect) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if errors.Is(err, playereffect.ErrEffectNotFound) {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return err
}
