// Package created contains the subscription created event.
package created

import (
	"time"

	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies a newly created subscription.
	Name bus.Name = "subscription.created"
)

// Payload contains created subscription state.
type Payload struct {
	// PlayerID identifies the member.
	PlayerID int64
	// Level stores the granted tier.
	Level record.Level
	// ExpiresAt stores the entitlement expiration.
	ExpiresAt time.Time
}
