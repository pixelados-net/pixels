// Package tick contains runtime ticking contracts.
package tick

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrInvalidInterval reports a missing or negative tick interval.
	ErrInvalidInterval = errors.New("invalid tick interval")

	// ErrInvalidTarget reports a missing tick target.
	ErrInvalidTarget = errors.New("invalid tick target")
)

// Tick describes one runtime tick.
type Tick struct {
	// At stores the tick time.
	At time.Time

	// Delta stores elapsed time since the previous tick.
	Delta time.Duration

	// Sequence stores the monotonic tick number.
	Sequence uint64
}

// Target handles ticks.
type Target interface {
	// Tick handles one runtime tick.
	Tick(context.Context, Tick) error
}

// TargetFunc adapts a function to a tick target.
type TargetFunc func(context.Context, Tick) error

// Tick handles one runtime tick.
func (target TargetFunc) Tick(ctx context.Context, tick Tick) error {
	return target(ctx, tick)
}

// Clock returns the current time.
type Clock interface {
	// Now returns the current time.
	Now() time.Time
}

// Scheduler controls a ticking lifecycle.
type Scheduler interface {
	// Start starts ticking the target.
	Start(context.Context, Target) error

	// Stop stops ticking.
	Stop() error

	// Done closes when ticking stops.
	Done() <-chan struct{}
}

// Config configures a ticker implementation.
type Config struct {
	// Interval stores the duration between ticks.
	Interval time.Duration
}

// Validate reports invalid ticker configuration.
func (config Config) Validate() error {
	if config.Interval <= 0 {
		return ErrInvalidInterval
	}

	return nil
}
