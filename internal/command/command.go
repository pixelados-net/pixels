// Package command contains small runtime command contracts.
package command

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrInvalidName reports a command without a name.
	ErrInvalidName = errors.New("invalid command name")

	// ErrInvalidHandler reports a missing command handler.
	ErrInvalidHandler = errors.New("invalid command handler")
)

// Name identifies a command intent.
type Name string

// Command describes a typed runtime command.
type Command interface {
	// CommandName returns the stable command name.
	CommandName() Name
}

// Metadata describes command routing context.
type Metadata struct {
	// PlayerID identifies the player that caused the command when available.
	PlayerID int64

	// ConnectionID identifies the source connection when available.
	ConnectionID string

	// CreatedAt stores when the command was created.
	CreatedAt time.Time
}

// Envelope wraps a command with runtime metadata.
type Envelope[T Command] struct {
	// Command stores the typed command.
	Command T

	// Metadata stores runtime routing context.
	Metadata Metadata
}

// Handler handles one typed command.
type Handler[T Command] interface {
	// Handle handles a command envelope.
	Handle(context.Context, Envelope[T]) error
}

// HandlerFunc adapts a function into a command handler.
type HandlerFunc[T Command] func(context.Context, Envelope[T]) error

// Handle handles a command envelope.
func (handler HandlerFunc[T]) Handle(ctx context.Context, envelope Envelope[T]) error {
	return handler(ctx, envelope)
}

// Middleware wraps a command handler.
type Middleware[T Command] func(Handler[T]) Handler[T]

// Chain applies middleware in the order it is provided.
func Chain[T Command](handler Handler[T], middleware ...Middleware[T]) Handler[T] {
	for index := len(middleware) - 1; index >= 0; index-- {
		handler = middleware[index](handler)
	}

	return handler
}

// Valid reports whether the envelope contains a named command.
func (envelope Envelope[T]) Valid() bool {
	return envelope.Command.CommandName() != ""
}

// WithCreatedAt returns the envelope with a default creation time when missing.
func (envelope Envelope[T]) WithCreatedAt(now time.Time) Envelope[T] {
	if envelope.Metadata.CreatedAt.IsZero() {
		envelope.Metadata.CreatedAt = now
	}

	return envelope
}
