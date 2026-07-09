// Package model contains inventory currency records.
package model

import "time"

// Balance stores one player's amount of one currency type.
type Balance struct {
	// PlayerID identifies the balance owner.
	PlayerID int64

	// CurrencyType identifies the protocol currency type.
	CurrencyType int32

	// Amount stores the non-negative balance.
	Amount int64

	// UpdatedAt stores the last balance mutation time.
	UpdatedAt time.Time

	// Version stores the optimistic record version.
	Version int64
}
