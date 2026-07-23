package routes

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	progressionpolicy "github.com/niflaot/pixels/internal/realm/progression/policy"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	progressionrequest "github.com/niflaot/pixels/pkg/http/progression/routes/request"
)

// PromoRequest aliases the promotional badge payload.
type PromoRequest = progressionrequest.Promo

// PromoClaimRequest aliases the manual promotion claim payload.
type PromoClaimRequest = progressionrequest.PromoClaim

// registerPromos mounts promotional badge administration.
func registerPromos(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath+"/promos", dependencies.listPromos)
	app.Post(basePath+"/promos", dependencies.createPromo)
	app.Patch(basePath+"/promos/:code", dependencies.updatePromo)
	app.Delete(basePath+"/promos/:code", dependencies.disablePromo)
	app.Get(basePath+"/promos/:code/claims", dependencies.promoClaims)
	app.Post(basePath+"/players/:playerId/promos/:code/claim", dependencies.claimPlayerPromo)
}

// listPromos returns the current immutable promotion catalog.
func (dependencies Dependencies) listPromos(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.ManageQuests); err != nil {
		return err
	}
	generation := dependencies.Catalog.Current()
	if generation == nil {
		return ctx.JSON([]progressionrecord.PromoBadge{})
	}
	return ctx.JSON(generation.Catalog.Promos)
}

// createPromo inserts one promotion without hot-reloading the catalog.
func (dependencies Dependencies) createPromo(ctx *fiber.Ctx) error {
	var request PromoRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	value := request.Value(request.Code)
	if !progressionrequest.ValidPromo(value) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid promotion")
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageQuests, request.Audit, "promo.create", entityID("promo", value.Code), func(txCtx context.Context) error { return dependencies.Admin.CreatePromo(txCtx, value) })
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(value)
}

// updatePromo replaces one promotional badge definition.
func (dependencies Dependencies) updatePromo(ctx *fiber.Ctx) error {
	var request PromoRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	value := request.Value(ctx.Params("code"))
	if !progressionrequest.ValidPromo(value) {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid promotion")
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageQuests, request.Audit, "promo.update", entityID("promo", value.Code), func(txCtx context.Context) error {
		changed, updateErr := dependencies.Admin.UpdatePromo(txCtx, value)
		if updateErr == nil && !changed {
			return progressionrecord.ErrNotFound
		}
		return updateErr
	})
	if err != nil {
		return err
	}
	return ctx.JSON(value)
}

// disablePromo soft-disables one promotional badge definition.
func (dependencies Dependencies) disablePromo(ctx *fiber.Ctx) error {
	code := strings.ToUpper(strings.TrimSpace(ctx.Params("code")))
	var request AuditRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	err := dependencies.mutate(ctx, progressionpolicy.ManageQuests, request, "promo.disable", entityID("promo", code), func(txCtx context.Context) error {
		changed, disableErr := dependencies.Admin.DisablePromo(txCtx, code)
		if disableErr == nil && !changed {
			return progressionrecord.ErrNotFound
		}
		return disableErr
	})
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// promoClaims lists durable claimants for one promotion.
func (dependencies Dependencies) promoClaims(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.ManageQuests); err != nil {
		return err
	}
	values, err := dependencies.Admin.PromoClaims(ctx.Context(), strings.ToUpper(strings.TrimSpace(ctx.Params("code"))))
	if err != nil {
		return err
	}
	return ctx.JSON(values)
}

// claimPlayerPromo grants one promotion through the real claim workflow.
func (dependencies Dependencies) claimPlayerPromo(ctx *fiber.Ctx) error {
	playerID, err := parsePositiveID(ctx, "playerId")
	if err != nil {
		return err
	}
	code := strings.ToUpper(strings.TrimSpace(ctx.Params("code")))
	var request PromoClaimRequest
	if err = parseBody(ctx, &request); err != nil {
		return err
	}
	claimed := false
	err = dependencies.mutate(ctx, progressionpolicy.OverridePlayers, request.Audit, "player.promo.claim", entityID("player-promo", playerID), func(txCtx context.Context) error {
		var claimErr error
		claimed, claimErr = dependencies.Promos.Claim(txCtx, playerID, code, request.Force)
		if claimErr == nil && !claimed {
			return progressionrecord.ErrConflict
		}
		return claimErr
	})
	if err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"claimed": claimed})
}
