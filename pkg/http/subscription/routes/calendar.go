package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// listCampaigns lists calendar campaigns.
func listCampaigns(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		campaigns, err := dependencies.Subscriptions.Campaigns(ctx.Context())
		if err != nil {
			return err
		}
		return ctx.JSON(campaigns)
	}
}

// saveCampaign creates or updates one campaign.
func saveCampaign(dependencies Dependencies, update bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request CampaignRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid calendar campaign request")
		}
		id := int64(0)
		if update {
			var err error
			id, err = identifier(ctx, "id")
			if err != nil {
				return err
			}
		}
		campaign, err := dependencies.Subscriptions.SaveCampaign(ctx.Context(), record.Campaign{ID: id, Name: request.Name,
			Image: request.Image, StartDate: request.StartDate, DayCount: request.DayCount, Enabled: request.Enabled}, request.Days)
		if err != nil {
			return routeError(err)
		}
		return ctx.JSON(campaign)
	}
}
