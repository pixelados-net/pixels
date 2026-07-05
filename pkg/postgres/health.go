package postgres

import (
	"context"
	"time"
)

// Pinger describes PostgreSQL ping behavior.
type Pinger interface {
	// Ping verifies connectivity.
	Ping(ctx context.Context) error
}

// Health verifies PostgreSQL availability.
type Health struct {
	// Pinger is the PostgreSQL connectivity checker.
	Pinger Pinger

	// Timeout is the maximum duration for health checks.
	Timeout time.Duration
}

// Ping verifies the PostgreSQL connection.
func (health Health) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, health.Timeout)
	defer cancel()

	return health.Pinger.Ping(ctx)
}
