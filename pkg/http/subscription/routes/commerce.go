package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// listClubOffers lists club offers.
func listClubOffers(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		offers, err := dependencies.Subscriptions.ClubOffers(ctx.Context())
		if err != nil {
			return err
		}
		return ctx.JSON(offers)
	}
}

// saveClubOffer creates or updates one club offer.
func saveClubOffer(dependencies Dependencies, update bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request ClubOfferRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid club offer request")
		}
		id := int64(0)
		if update {
			var err error
			id, err = identifier(ctx, "id")
			if err != nil {
				return err
			}
		}
		offer, err := dependencies.Subscriptions.SaveClubOffer(ctx.Context(), record.Offer{ID: id, Name: request.Name,
			DayCount: request.DayCount, PriceCredits: request.PriceCredits, PricePoints: request.PricePoints,
			PointsType: request.PointsType, VIP: request.VIP, Deal: request.Deal, Enabled: request.Enabled, OrderNum: request.OrderNum})
		if err != nil {
			return routeError(err)
		}
		return ctx.JSON(offer)
	}
}

// listTargetedOffers lists targeted offers.
func listTargetedOffers(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		offers, err := dependencies.Subscriptions.TargetedOffers(ctx.Context())
		if err != nil {
			return err
		}
		return ctx.JSON(offers)
	}
}

// saveTargetedOffer creates or updates one targeted offer.
func saveTargetedOffer(dependencies Dependencies, update bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request TargetedOfferRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid targeted offer request")
		}
		id := int64(0)
		if update {
			var err error
			id, err = identifier(ctx, "id")
			if err != nil {
				return err
			}
		}
		offer, err := dependencies.Subscriptions.SaveTargetedOffer(ctx.Context(), record.TargetedOffer{ID: id,
			CatalogItemID: request.CatalogItemID, PriceCredits: request.PriceCredits, PricePoints: request.PricePoints,
			PointsType: request.PointsType, PurchaseLimit: request.PurchaseLimit, TitleKey: request.TitleKey,
			DescriptionKey: request.DescriptionKey, ImageURL: request.ImageURL, IconURL: request.IconURL,
			ExpiresAt: request.ExpiresAt, OrderNum: request.OrderNum, Enabled: request.Enabled})
		if err != nil {
			return routeError(err)
		}
		return ctx.JSON(offer)
	}
}
