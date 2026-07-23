// Package redeemed contains the committed exchange event.
package redeemed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one committed exchange.
const Name bus.Name = "exchange.redeemed"

// Payload stores bounded exchange identifiers and amount.
type Payload struct {
	PlayerID int64
	ItemID   int64
	Credits  int64
}
