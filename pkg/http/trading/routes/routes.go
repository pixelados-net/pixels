// Package routes exposes protected Marketplace and direct-trade administration.
package routes

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	tradeadmin "github.com/niflaot/pixels/internal/realm/trade/admin"
	"go.uber.org/fx"
)

// AuditResponse contains one completed direct trade.
type AuditResponse struct {
	// ID identifies the audit row.
	ID int64 `json:"id"`
	// RoomID identifies the settlement room.
	RoomID int64 `json:"roomId"`
	// FirstPlayerID identifies the first participant.
	FirstPlayerID int64 `json:"firstPlayerId"`
	// SecondPlayerID identifies the second participant.
	SecondPlayerID int64 `json:"secondPlayerId"`
	// FirstIP stores the first participant audit address.
	FirstIP string `json:"firstIp,omitempty"`
	// SecondIP stores the second participant audit address.
	SecondIP string `json:"secondIp,omitempty"`
	// FirstItemIDs stores the first offer.
	FirstItemIDs []int64 `json:"firstItemIds"`
	// SecondItemIDs stores the second offer.
	SecondItemIDs []int64 `json:"secondItemIds"`
	// FirstRedeemableCredits stores value delivered to the second participant.
	FirstRedeemableCredits int64 `json:"firstRedeemableCredits"`
	// SecondRedeemableCredits stores value delivered to the first participant.
	SecondRedeemableCredits int64 `json:"secondRedeemableCredits"`
	// CreatedAt stores settlement time.
	CreatedAt time.Time `json:"createdAt"`
}

// Dependencies contains trading administration behavior.
type Dependencies struct {
	// In marks dependencies for Fx injection.
	fx.In
	// Marketplace manages protected listing lifecycle operations.
	Marketplace *marketcore.Service
	// Trade manages locks and audit reads.
	Trade *tradeadmin.Service
	// Sanctions owns the superseding global trade-lock records.
	Sanctions *sanctioncore.Service
}

// Register mounts protected trading administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	if dependencies.Trade != nil {
		app.Get("/api/admin/trade/players/:playerId/log", dependencies.logs)
		app.Post("/api/admin/trade/players/:playerId/lock", dependencies.lock)
		app.Delete("/api/admin/trade/players/:playerId/lock", dependencies.unlock)
	}
	if dependencies.Marketplace != nil {
		app.Post("/api/admin/marketplace/listings/:id/force-close", dependencies.forceClose)
	}
}

// playerID parses a positive path identifier.
func playerID(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid path identifier")
	}
	return value, nil
}

// logs returns recent completed trades involving a player.
func (dependencies Dependencies) logs(ctx *fiber.Ctx) error {
	id, err := playerID(ctx, "playerId")
	if err != nil {
		return err
	}
	values, err := dependencies.Trade.Logs(ctx.Context(), id)
	if err != nil {
		return err
	}
	items := make([]AuditResponse, len(values))
	for index, value := range values {
		items[index] = AuditResponse{ID: value.ID, RoomID: value.RoomID, FirstPlayerID: value.FirstPlayerID, SecondPlayerID: value.SecondPlayerID, FirstIP: value.FirstIP, SecondIP: value.SecondIP, FirstItemIDs: value.FirstItemIDs, SecondItemIDs: value.SecondItemIDs, FirstRedeemableCredits: value.FirstRedeemableCredits, SecondRedeemableCredits: value.SecondRedeemableCredits, CreatedAt: value.CreatedAt}
	}
	return ctx.JSON(fiber.Map{"items": items, "count": len(items)})
}

// lock disables direct trading for a player.
func (dependencies Dependencies) lock(ctx *fiber.Ctx) error { return dependencies.setLock(ctx, true) }

// unlock enables direct trading for a player.
func (dependencies Dependencies) unlock(ctx *fiber.Ctx) error {
	return dependencies.setLock(ctx, false)
}

// setLock applies one durable trade lock.
func (dependencies Dependencies) setLock(ctx *fiber.Ctx, locked bool) error {
	id, err := playerID(ctx, "playerId")
	if err != nil {
		return err
	}
	if dependencies.Sanctions == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "global sanction service unavailable")
	}
	if locked {
		_, err = dependencies.Sanctions.Apply(ctx.Context(), sanctionrecord.ApplyParams{ReceiverPlayerID: id, IssuerKind: "system", Kind: sanctionrecord.KindTradeLock, Reason: "Administrative direct-trade lock", Source: "admin_http"})
		if err != nil {
			return err
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
	history, err := dependencies.Sanctions.History(ctx.Context(), id, 500)
	if err != nil {
		return err
	}
	for _, punishment := range history {
		if punishment.Kind == sanctionrecord.KindTradeLock && punishment.ActiveAt(time.Now()) && (punishment.Source == "admin_http" || strings.HasPrefix(punishment.Reason, "Migrated legacy")) {
			if _, err = dependencies.Sanctions.RevokeSystem(ctx.Context(), punishment.ID); err != nil {
				return err
			}
		}
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// forceClose returns one open Marketplace item to its seller.
func (dependencies Dependencies) forceClose(ctx *fiber.Ctx) error {
	id, err := playerID(ctx, "id")
	if err != nil {
		return err
	}
	err = dependencies.Marketplace.Close(ctx.Context(), id, 0, true)
	if errors.Is(err, marketcore.ErrListingUnavailable) {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
