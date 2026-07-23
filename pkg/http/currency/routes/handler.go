package routes

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// walletHandler returns one persistent player's configured wallet.
func walletHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := playerIDQuery(ctx)
		if err != nil {
			return err
		}
		if err := requirePlayer(ctx.Context(), dependencies, playerID); err != nil {
			return err
		}

		balances, err := dependencies.Currencies.Wallet(ctx.Context(), playerID)
		if err != nil {
			return fmt.Errorf("read player %d currency wallet: %w", playerID, err)
		}

		return ctx.JSON(WalletResponse{PlayerID: playerID, Balances: balanceResponses(balances)})
	}
}

// typesHandler returns configured currency definitions.
func typesHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		definitions, err := dependencies.Currencies.Types(ctx.Context())
		if err != nil {
			return fmt.Errorf("read currency types: %w", err)
		}

		return ctx.JSON(TypesResponse{Types: typeResponses(definitions)})
	}
}

// requirePlayer verifies a persistent player identity.
func requirePlayer(ctx context.Context, dependencies Dependencies, playerID int64) error {
	_, found, err := dependencies.Finder.FindByID(ctx, playerID)
	if err != nil {
		return fmt.Errorf("find currency player %d: %w", playerID, err)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "currency player not found")
	}

	return nil
}
