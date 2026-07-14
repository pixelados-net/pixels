package effect

import (
	"context"
	"time"
)

// Store persists player effects and their selected projection.
type Store interface {
	// WithinTransaction runs work in one shared transaction.
	WithinTransaction(context.Context, func(context.Context) error) error
	// List returns one player's durable effects.
	List(context.Context, int64) ([]Effect, error)
	// Grant creates or increments one effect stack.
	Grant(context.Context, int64, int32, int32) (Effect, error)
	// Activate starts one available effect charge.
	Activate(context.Context, int64, int32, time.Time) (Effect, bool, error)
	// SetActive replaces the selected effect id.
	SetActive(context.Context, int64, *int32) error
	// Active returns one player's selected effect id.
	Active(context.Context, int64) (*int32, error)
	// Revoke deletes one effect stack.
	Revoke(context.Context, int64, int32) (bool, error)
	// Expire consumes expired active charges in one bounded batch.
	Expire(context.Context, time.Time, int32) ([]Expiration, error)
}

// Manager exposes effect behavior to catalog, furniture, and HTTP adapters.
type Manager interface {
	// List returns the effective durable and synthetic inventory.
	List(context.Context, int64) ([]Effect, error)
	// Grant adds one durable charge.
	Grant(context.Context, int64, int32, int32, Source) (Effect, error)
	// Enable activates and selects one effect, or disables it with id zero.
	Enable(context.Context, int64, int32) error
	// Activate starts one effect charge without selecting it.
	Activate(context.Context, int64, int32) (Effect, error)
	// Revoke removes one durable effect.
	Revoke(context.Context, int64, int32) error
}
