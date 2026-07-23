package model

import "time"

// LedgerEntry stores one audited currency balance mutation.
type LedgerEntry struct {
	// ID identifies the ledger entry.
	ID int64

	// PlayerID identifies the affected player.
	PlayerID int64

	// CurrencyType identifies the affected currency.
	CurrencyType int32

	// Delta stores the signed balance change.
	Delta int64

	// BalanceAfter stores the resulting absolute balance.
	BalanceAfter int64

	// Reason stores the mutation audit reason.
	Reason string

	// ActorKind identifies the mutation source family.
	ActorKind string

	// ActorID optionally identifies the mutation source.
	ActorID *int64

	// CreatedAt stores when the mutation was recorded.
	CreatedAt time.Time
}
