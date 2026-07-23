package service

import "errors"

var (
	// ErrInvalidPlayerID reports a non-positive player id.
	ErrInvalidPlayerID = errors.New("invalid currency player id")

	// ErrInvalidCurrencyType reports a currency absent from the catalog.
	ErrInvalidCurrencyType = errors.New("invalid currency type")

	// ErrInsufficientBalance reports a deduction that would make a balance negative.
	ErrInsufficientBalance = errors.New("insufficient currency balance")

	// ErrInvalidAmount reports a zero grant or negative absolute balance.
	ErrInvalidAmount = errors.New("invalid currency amount")

	// ErrInvalidActor reports an unsupported mutation source.
	ErrInvalidActor = errors.New("invalid currency actor")
)
