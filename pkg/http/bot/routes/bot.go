package routes

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

// readBot returns durable bot support state.
func (dependencies Dependencies) readBot(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return err
	}
	bot, found, err := dependencies.Bots.Find(ctx.Context(), id)
	if err != nil {
		return err
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, botrecord.ErrBotNotFound.Error())
	}
	return ctx.JSON(botResponse(bot))
}

// forcePickup administratively returns a placed bot to its owner.
func (dependencies Dependencies) forcePickup(ctx *fiber.Ctx) error {
	id, err := positiveID(ctx.Params("id"))
	if err != nil {
		return err
	}
	bot, err := dependencies.Bots.ForcePickup(ctx.Context(), id)
	if errors.Is(err, botrecord.ErrBotNotFound) {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	if errors.Is(err, botrecord.ErrConflict) {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}
	if err != nil {
		return err
	}
	return ctx.JSON(botResponse(bot))
}

// botResponse maps complete durable bot support state.
func botResponse(bot botrecord.Bot) BotResponse {
	return BotResponse{ID: bot.ID, OwnerPlayerID: bot.OwnerPlayerID, OwnerName: bot.OwnerName, RoomID: bot.RoomID, BehaviorType: bot.BehaviorType, Name: bot.Name, Motto: bot.Motto, Figure: bot.Figure, Gender: bot.Gender, X: bot.X, Y: bot.Y, Z: bot.Z, Rotation: bot.Rotation, CanWalk: bot.CanWalk, DanceType: bot.DanceType, ChatAuto: bot.ChatAuto, ChatRandom: bot.ChatRandom, ChatDelaySeconds: bot.ChatDelaySeconds, BubbleStyle: bot.BubbleStyle, EffectID: bot.EffectID, ChatLines: bot.ChatLines, CreatedAt: bot.CreatedAt, UpdatedAt: bot.UpdatedAt, Version: bot.Version}
}
