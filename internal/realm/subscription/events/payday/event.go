// Package payday contains the subscription payday event.
package payday

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies one committed kickback award.
	Name bus.Name = "subscription.payday.awarded"
)

// Payload contains one payday reward.
type Payload struct {
	// PlayerID identifies the beneficiary.
	PlayerID int64
	// Amount stores the awarded balance.
	Amount int64
	// CurrencyType identifies the reward currency.
	CurrencyType int32
	// Streak stores active membership days.
	Streak int32
}
