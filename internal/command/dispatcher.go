package command

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Dispatcher validates and dispatches one typed command kind.
type Dispatcher[T Command] struct {
	// handler receives valid command envelopes.
	handler Handler[T]
	// log records dispatched commands.
	log *zap.Logger
}

// NewDispatcher creates a command dispatcher.
func NewDispatcher[T Command](handler Handler[T], middleware ...Middleware[T]) (*Dispatcher[T], error) {
	if handler == nil {
		return nil, ErrInvalidHandler
	}

	return &Dispatcher[T]{handler: Chain(handler, middleware...), log: zap.NewNop()}, nil
}

// WithLogger configures structured command dispatch logging.
func (dispatcher *Dispatcher[T]) WithLogger(log *zap.Logger) *Dispatcher[T] {
	if log == nil {
		log = zap.NewNop()
	}

	dispatcher.log = log

	return dispatcher
}

// Dispatch sends an envelope to the configured handler.
func (dispatcher *Dispatcher[T]) Dispatch(ctx context.Context, envelope Envelope[T]) error {
	envelope = envelope.WithCreatedAt(time.Now())
	if !envelope.Valid() {
		return ErrInvalidName
	}

	dispatcher.log.Debug("command dispatched",
		zap.String("command_name", string(envelope.Command.CommandName())),
		zap.Int64("player_id", envelope.Metadata.PlayerID),
		zap.String("connection_id", envelope.Metadata.ConnectionID),
		zap.Time("created_at", envelope.Metadata.CreatedAt),
		zap.Any("command", envelope.Command),
	)

	return dispatcher.handler.Handle(ctx, envelope)
}
