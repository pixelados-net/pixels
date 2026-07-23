// Package expired contains the subscription expired event.
package expired

import (
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies an expired subscription.
	Name bus.Name = "subscription.expired"
)

// Payload contains expired subscription state.
type Payload struct {
	// PlayerID identifies the former member.
	PlayerID int64
	// PreviousLevel stores the expired tier.
	PreviousLevel record.Level
}
