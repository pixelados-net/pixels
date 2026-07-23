// Package changed contains the inventory currency changed event.
package changed

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies a committed player currency mutation.
	Name bus.Name = "inventory.currency_changed"
)

// Payload describes one committed player currency mutation.
type Payload struct {
	// PlayerID identifies the affected player.
	PlayerID int64

	// CurrencyType identifies the affected protocol currency.
	CurrencyType int32

	// Amount stores the resulting absolute balance.
	Amount int64

	// Delta stores the signed committed balance change.
	Delta int64

	// ActorKind identifies the mutation source family.
	ActorKind string
}
