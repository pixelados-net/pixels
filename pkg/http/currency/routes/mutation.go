package routes

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
)

// mutationAction names one administrative balance operation.
type mutationAction string

const (
	// grantAction adds a positive amount.
	grantAction mutationAction = "grant"

	// deductAction subtracts a positive amount.
	deductAction mutationAction = "deduct"

	// setAction replaces the absolute balance.
	setAction mutationAction = "set"
)

// mutationHandler commits one administrative currency mutation.
func mutationHandler(action mutationAction, dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		input, err := parseMutationInput(ctx, action)
		if err != nil {
			return err
		}
		if err := requirePlayer(ctx.Context(), dependencies, input.playerID); err != nil {
			return err
		}

		amount, err := action.apply(ctx, dependencies, input)
		if err != nil {
			return mutationError(input, err)
		}
		alertSent := sendMutationAlert(ctx, dependencies, action, input, amount)

		return ctx.JSON(MutationResponse{
			PlayerID: input.playerID, CurrencyType: input.currencyType, Amount: amount,
			AlertRequested: input.request.Alert, AlertSent: alertSent,
		})
	}
}

// validate validates action-specific request amounts.
func (action mutationAction) validate(amount int64) error {
	if action == setAction {
		if amount < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "currency set amount must be non-negative")
		}
		return nil
	}
	if amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "currency mutation amount must be positive")
	}

	return nil
}

// apply executes one action through currency management behavior.
func (action mutationAction) apply(ctx *fiber.Ctx, dependencies Dependencies, input mutationInput) (int64, error) {
	reason := input.request.Reason
	if reason == "" {
		reason = "admin_api_" + string(action)
	}
	if action == setAction {
		return dependencies.Currencies.Set(ctx.Context(), currencyservice.SetParams{
			PlayerID: input.playerID, CurrencyType: input.currencyType, Amount: input.request.Amount,
			Reason: reason, ActorKind: currencyservice.ActorAdmin,
		})
	}

	delta := input.request.Amount
	if action == deductAction {
		delta = -delta
	}

	return dependencies.Currencies.Grant(ctx.Context(), currencyservice.GrantParams{
		PlayerID: input.playerID, CurrencyType: input.currencyType, Amount: delta,
		Reason: reason, ActorKind: currencyservice.ActorAdmin,
	})
}

// mutationError maps domain errors into meaningful HTTP failures.
func mutationError(input mutationInput, err error) error {
	switch {
	case errors.Is(err, currencyservice.ErrInvalidCurrencyType):
		return fiber.NewError(fiber.StatusBadRequest, "currency type is not configured")
	case errors.Is(err, currencyservice.ErrInvalidAmount), errors.Is(err, currencyservice.ErrInvalidActor):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, currencyservice.ErrInsufficientBalance):
		return fiber.NewError(fiber.StatusConflict, "currency deduction exceeds player balance")
	default:
		return fmt.Errorf("mutate player %d currency %d: %w", input.playerID, input.currencyType, err)
	}
}
