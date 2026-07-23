// Package extended contains the subscription extended event.
package extended

import (
	"time"

	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies an extended subscription.
	Name bus.Name = "subscription.extended"
)

// Payload contains extended subscription state.
type Payload struct {
	// PlayerID identifies the member.
	PlayerID int64
	// Level stores the resulting tier.
	Level record.Level
	// ExpiresAt stores the new expiration.
	ExpiresAt time.Time
}
