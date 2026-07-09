package routes

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// MutationRequest contains one administrative currency mutation.
type MutationRequest struct {
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

// parseMutationInput parses and validates mutation path and body input.
func parseMutationInput(ctx *fiber.Ctx, action mutationAction) (mutationInput, error) {
	playerID, err := positiveInt64(ctx.Params("id"), "player id")
	if err != nil {
		return mutationInput{}, err
	}
	currencyType, err := int32Value(ctx.Params("type"), "currency type")
	if err != nil {
		return mutationInput{}, err
	}

	var request MutationRequest
	if err := ctx.BodyParser(&request); err != nil {
		return mutationInput{}, fiber.NewError(fiber.StatusBadRequest, "invalid currency mutation request body")
	}
	if err := action.validate(request.Amount); err != nil {
		return mutationInput{}, err
	}
	request.Reason = strings.TrimSpace(request.Reason)

	return mutationInput{playerID: playerID, currencyType: currencyType, request: request}, nil
}

// playerIDParam parses a positive player id path parameter.
func playerIDParam(ctx *fiber.Ctx) (int64, error) {
	return positiveInt64(ctx.Params("id"), "player id")
}

// positiveInt64 parses one positive signed integer.
func positiveInt64(value string, name string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}

	return parsed, nil
}

// int32Value parses one signed protocol integer.
func int32Value(value string, name string) (int32, error) {
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}

	return int32(parsed), nil
}
