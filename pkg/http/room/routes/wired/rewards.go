package wired

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
)

// rewards returns normalized reward options without claim history.
func (handler Handler) rewards(ctx *fiber.Ctx) error {
	_, itemID, err := handler.identifiers(ctx)
	if err != nil {
		return err
	}
	items, err := handler.dependencies.Rewards.ListRewards(ctx.Context(), itemID)
	if err != nil {
		return err
	}
	return ctx.JSON(items)
}

// replaceRewards atomically replaces normalized reward options.
func (handler Handler) replaceRewards(ctx *fiber.Ctx) error {
	roomID, itemID, err := handler.identifiers(ctx)
	if err != nil {
		return err
	}
	if _, found, findErr := handler.dependencies.Store.Find(ctx.Context(), roomID, itemID); findErr != nil || !found {
		return handler.findError(findErr)
	}
	var request RewardsRequest
	if err = ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid WIRED rewards request")
	}
	items := make([]record.Reward, len(request.Items))
	for index, item := range request.Items {
		if !rewardKind(item.Kind) || item.Reference == "" || item.Amount <= 0 || item.Weight <= 0 || item.Stock != nil && *item.Stock < 0 {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid WIRED reward")
		}
		items[index] = record.Reward{Ordinal: index, Kind: item.Kind, Reference: item.Reference, Amount: item.Amount, Weight: item.Weight, Stock: item.Stock}
	}
	if err = handler.dependencies.Rewards.ReplaceRewards(ctx.Context(), itemID, items); err != nil {
		return persistenceError(err)
	}
	return ctx.JSON(ActionResponse{Success: true})
}

// rewardKind reports whether a reward uses one implemented durable capability.
func rewardKind(kind string) bool {
	switch kind {
	case "furniture", "badge", "credits", "currency", "respect", "catalog_offer":
		return true
	default:
		return false
	}
}

// deleteRewards clears all normalized reward options.
func (handler Handler) deleteRewards(ctx *fiber.Ctx) error {
	_, itemID, err := handler.identifiers(ctx)
	if err != nil {
		return err
	}
	if err = handler.dependencies.Rewards.ReplaceRewards(ctx.Context(), itemID, nil); err != nil {
		return persistenceError(err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
