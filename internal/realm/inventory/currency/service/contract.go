// Package service contains inventory currency behavior.
package service

import (
	"context"

	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
)

const (
	// ActorSystem identifies an automated server mutation.
	ActorSystem = "system"

	// ActorAdmin identifies an administrative mutation.
	ActorAdmin = "admin"

	// ActorPlayer identifies a player-originated mutation.
	ActorPlayer = "player"
)

// Reader reads player currency balances and configured types.
type Reader interface {
	// Wallet returns every configured currency balance for a player.
	Wallet(ctx context.Context, playerID int64) ([]currencymodel.Balance, error)

	// Balance returns one configured currency balance.
	Balance(ctx context.Context, playerID int64, currencyType int32) (int64, error)

	// Types returns configured currency definitions.
	Types(ctx context.Context) ([]currencymodel.Definition, error)
}

// Granter applies signed player currency mutations.
type Granter interface {
	// Grant applies a signed currency delta.
	Grant(ctx context.Context, params GrantParams) (int64, error)
}

// Manager reads and mutates player currency balances.
type Manager interface {
	Reader
	Granter

	// Set replaces a currency balance with an absolute amount.
	Set(ctx context.Context, params SetParams) (int64, error)
}

// GrantParams contains a signed currency balance change.
type GrantParams struct {
	// PlayerID identifies the affected player.
	PlayerID int64

	// CurrencyType identifies the affected currency.
	CurrencyType int32

	// Amount stores the signed balance delta.
	Amount int64

	// Reason stores the audit reason.
	Reason string

	// ActorKind identifies the mutation source family.
	ActorKind string

	// ActorID optionally identifies the mutation source.
	ActorID *int64
}

// SetParams contains an absolute currency balance correction.
type SetParams struct {
	// PlayerID identifies the affected player.
	PlayerID int64

	// CurrencyType identifies the affected currency.
	CurrencyType int32

	// Amount stores the new absolute balance.
	Amount int64

	// Reason stores the audit reason.
	Reason string

	// ActorKind identifies the mutation source family.
	ActorKind string

	// ActorID optionally identifies the mutation source.
	ActorID *int64
}
