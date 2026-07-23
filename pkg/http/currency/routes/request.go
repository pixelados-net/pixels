package routes

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// MutationRequest contains one administrative currency mutation.
type MutationRequest struct {
	// PlayerID identifies the target player.
	PlayerID int64 `json:"playerId"`

	// CurrencyType identifies the target protocol currency.
	CurrencyType *int32 `json:"currencyType"`

	// Amount stores a positive delta or non-negative absolute balance.
	Amount int64 `json:"amount"`

	// Reason stores an optional ledger audit reason.
	Reason string `json:"reason,omitempty"`

	// Alert requests an additional localized generic alert.
	Alert bool `json:"alert,omitempty"`

	// Locale optionally overrides the alert locale.
	Locale string `json:"locale,omitempty"`
}

// mutationInput contains parsed mutation route input.
type mutationInput struct {
	// playerID identifies the target player.
	playerID int64

	// currencyType identifies the target currency.
	currencyType int32

	// request stores the parsed request body.
	request MutationRequest
}

// parseMutationInput parses and validates currency mutation body input.
func parseMutationInput(ctx *fiber.Ctx, action mutationAction) (mutationInput, error) {
	var request MutationRequest
	if err := ctx.BodyParser(&request); err != nil {
		return mutationInput{}, fiber.NewError(fiber.StatusBadRequest, "invalid currency mutation request body")
	}
	if request.PlayerID <= 0 {
		return mutationInput{}, fiber.NewError(fiber.StatusBadRequest, "invalid player id")
	}
	if request.CurrencyType == nil {
		return mutationInput{}, fiber.NewError(fiber.StatusBadRequest, "currency type is required")
	}
	if err := action.validate(request.Amount); err != nil {
		return mutationInput{}, err
	}
	request.Reason = strings.TrimSpace(request.Reason)

	return mutationInput{playerID: request.PlayerID, currencyType: *request.CurrencyType, request: request}, nil
}

// playerIDQuery parses the required wallet player query parameter.
func playerIDQuery(ctx *fiber.Ctx) (int64, error) {
	return positiveInt64(ctx.Query("playerId"), "player id")
}

// positiveInt64 parses one positive signed integer.
func positiveInt64(value string, name string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}

	return parsed, nil
}
