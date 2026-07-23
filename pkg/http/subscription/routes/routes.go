// Package routes contains protected subscription administration routes.
package routes

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	subadmin "github.com/niflaot/pixels/internal/realm/subscription/admin"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	"go.uber.org/fx"
)

const (
	// basePath stores the subscription administration prefix.
	basePath = "/api/admin/subscriptions"
)

// Dependencies contains subscription administration collaborators.
type Dependencies struct {
	fx.In

	// Subscriptions manages administrative subscription behavior.
	Subscriptions *subadmin.Service
}

// Register registers protected subscription administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	if dependencies.Subscriptions == nil {
		return
	}
	app.Get(basePath+"/club-offers", listClubOffers(dependencies))
	app.Post(basePath+"/club-offers", saveClubOffer(dependencies, false))
	app.Patch(basePath+"/club-offers/:id", saveClubOffer(dependencies, true))
	app.Get(basePath+"/targeted-offers", listTargetedOffers(dependencies))
	app.Post(basePath+"/targeted-offers", saveTargetedOffer(dependencies, false))
	app.Patch(basePath+"/targeted-offers/:id", saveTargetedOffer(dependencies, true))
	app.Get(basePath+"/calendar/campaigns", listCampaigns(dependencies))
	app.Post(basePath+"/calendar/campaigns", saveCampaign(dependencies, false))
	app.Patch(basePath+"/calendar/campaigns/:id", saveCampaign(dependencies, true))
	app.Get(basePath+"/:playerId", membership(dependencies))
	app.Post(basePath+"/:playerId/grant", grant(dependencies))
	app.Delete(basePath+"/:playerId", revoke(dependencies))
}

// membership returns membership and payday history.
func membership(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := identifier(ctx, "playerId")
		if err != nil {
			return err
		}
		membership, payday, paydays, found, err := dependencies.Subscriptions.Membership(ctx.Context(), playerID)
		if err != nil {
			return err
		}
		if !found {
			return fiber.NewError(fiber.StatusNotFound, "subscription membership not found")
		}
		return ctx.JSON(fiber.Map{"membership": membership, "paydayProjection": payday,
			"giftsAvailable": core.RemainingClubGifts(membership), "paydays": paydays})
	}
}

// grant grants or extends one membership.
func grant(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := identifier(ctx, "playerId")
		if err != nil {
			return err
		}
		var request GrantRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid membership grant request")
		}
		membership, err := dependencies.Subscriptions.Grant(ctx.Context(), playerID, request.Level, time.Duration(request.DurationSeconds)*time.Second)
		if err != nil {
			return routeError(err)
		}
		return ctx.JSON(membership)
	}
}

// revoke revokes one membership.
func revoke(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := identifier(ctx, "playerId")
		if err != nil {
			return err
		}
		if err := dependencies.Subscriptions.Revoke(ctx.Context(), playerID); err != nil {
			return routeError(err)
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

// identifier parses one positive route identifier.
func identifier(ctx *fiber.Ctx, name string) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid subscription record id")
	}
	return id, nil
}

// routeError maps subscription administration failures.
func routeError(err error) error {
	if errors.Is(err, subadmin.ErrInvalidInput) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if errors.Is(err, subadmin.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return err
}
