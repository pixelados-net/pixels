package repository

import (
	"errors"

	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
)

var (
	// ErrInsufficientBalance reports a mutation that would create a negative balance.
	ErrInsufficientBalance = errors.New("insufficient currency balance")

	// ErrBalanceOverflow reports a mutation outside signed 64-bit storage.
	ErrBalanceOverflow = errors.New("currency balance overflow")
)

// Mutation contains one atomic balance and ledger mutation.
type Mutation struct {
	// PlayerID identifies the affected player.
	PlayerID int64

	// CurrencyType identifies the affected currency.
	CurrencyType int32

	// Amount stores a signed delta for Grant or an absolute value for Set.
	Amount int64

	// Ledger reports whether an audit entry must be written.
	Ledger bool

	// Reason stores the audit reason.
	Reason string

	// ActorKind identifies the mutation source family.
	ActorKind string

	// ActorID optionally identifies the mutation source.
	ActorID *int64
}

// Result contains the committed balance and its signed delta.
type Result struct {
	// Balance stores the committed currency balance.
	Balance currencymodel.Balance

	// Delta stores the signed change committed by the mutation.
	Delta int64
}

// ledgerEntry creates an audit record from a completed mutation.
func ledgerEntry(mutation Mutation, delta int64, balance int64) currencymodel.LedgerEntry {
	return currencymodel.LedgerEntry{
		PlayerID:     mutation.PlayerID,
		CurrencyType: mutation.CurrencyType,
		Delta:        delta,
		BalanceAfter: balance,
		Reason:       mutation.Reason,
		ActorKind:    mutation.ActorKind,
		ActorID:      mutation.ActorID,
	}
}
