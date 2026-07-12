// Package chatlog persists optional messenger chat outside the delivery path.
package chatlog

import (
	"context"
	"sync"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

const queueCapacity = 512

// Entry contains one accepted private message audit record.
type Entry struct {
	// FromPlayerID identifies the sender.
	FromPlayerID int64
	// ToPlayerID identifies the recipient.
	ToPlayerID int64
	// Message stores filtered message text.
	Message string
}

// Writer drains private-message records outside packet handling.
type Writer struct {
	// enabled reports whether persistence is configured.
	enabled bool
	// store persists queued records.
	store MessageStore
	// log records asynchronous persistence failures.
	log *zap.Logger
	// queue stores bounded pending records.
	queue chan Entry
	// done closes after the worker exits.
	done chan struct{}
	// once protects worker startup.
	once sync.Once
}

// New creates optional private-message persistence.
func New(config Config, store MessageStore, log *zap.Logger) *Writer {
	if log == nil {
		log = zap.NewNop()
	}
	return &Writer{enabled: config.Enabled, store: store, log: log, queue: make(chan Entry, queueCapacity), done: make(chan struct{})}
}

// MessageStore persists one private message record.
type MessageStore interface {
	// LogPrivateMessage persists one accepted private message.
	LogPrivateMessage(context.Context, int64, int64, string) error
}

// Config contains private-message writer settings.
type Config struct {
	// Enabled reports whether records should be persisted.
	Enabled bool
}

// Enqueue queues one record without blocking live packet delivery.
func (writer *Writer) Enqueue(fromID int64, toID int64, message string) bool {
	if writer == nil || !writer.enabled {
		return true
	}
	select {
	case writer.queue <- Entry{FromPlayerID: fromID, ToPlayerID: toID, Message: message}:
		return true
	default:
		writer.log.Warn("messenger private chat log queue full", zap.Int64("from_player_id", fromID), zap.Int64("to_player_id", toID))
		return false
	}
}

// Start starts the asynchronous writer once.
func (writer *Writer) Start(ctx context.Context) {
	writer.once.Do(func() {
		go writer.run(ctx)
	})
}

// Wait waits until the writer exits.
func (writer *Writer) Wait() { <-writer.done }

// run drains queued messages until shutdown.
func (writer *Writer) run(ctx context.Context) {
	defer close(writer.done)
	for {
		select {
		case <-ctx.Done():
			writer.drain()
			return
		case entry := <-writer.queue:
			if err := writer.store.LogPrivateMessage(context.Background(), entry.FromPlayerID, entry.ToPlayerID, entry.Message); err != nil {
				writer.log.Warn("messenger private chat log failed", zap.Int64("from_player_id", entry.FromPlayerID), zap.Int64("to_player_id", entry.ToPlayerID), zap.Error(err))
			}
		}
	}
}

// drain persists records already accepted before shutdown.
func (writer *Writer) drain() {
	for {
		select {
		case entry := <-writer.queue:
			if err := writer.store.LogPrivateMessage(context.Background(), entry.FromPlayerID, entry.ToPlayerID, entry.Message); err != nil {
				writer.log.Warn("messenger private chat drain failed", zap.Int64("from_player_id", entry.FromPlayerID), zap.Int64("to_player_id", entry.ToPlayerID), zap.Error(err))
			}
		default:
			return
		}
	}
}

// RegisterLifecycle starts and stops private-message persistence.
func RegisterLifecycle(lifecycle fx.Lifecycle, writer *Writer) {
	var cancel context.CancelFunc
	lifecycle.Append(fx.Hook{OnStart: func(context.Context) error {
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		writer.Start(ctx)
		return nil
	}, OnStop: func(context.Context) error {
		cancel()
		writer.Wait()
		return nil
	}})
}
