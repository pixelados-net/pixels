// Package history asynchronously persists admitted Navigator room visits.
package history

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	navrecord "github.com/niflaot/pixels/internal/realm/navigator/record"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Writer owns a bounded asynchronous room-visit queue.
type Writer struct {
	// navigator persists visit aggregates.
	navigator navservice.Manager
	// log records dropped or failed telemetry without breaking admission.
	log *zap.Logger
	// queue contains value-only visit events.
	queue chan roomentered.Payload
	// cancel stops the worker.
	cancel context.CancelFunc
	// wait tracks the worker lifecycle.
	wait sync.WaitGroup
	// duplicateWindow prevents rapid reentries from inflating visits.
	duplicateWindow time.Duration
	// now supplies a deterministic clock for deduplication tests.
	now func() time.Time
	// dropped counts telemetry rejected by backpressure.
	dropped atomic.Uint64
	// batchSize bounds one grouped database operation.
	batchSize int
}

// New creates a bounded room-visit writer.
func New(navigator navservice.Manager, log *zap.Logger) *Writer {
	return NewConfigured(navigator, log, 1024, 30*time.Second)
}

// NewConfigured creates a room-visit writer with explicit queue policy.
func NewConfigured(navigator navservice.Manager, log *zap.Logger, queueSize int, duplicateWindow time.Duration) *Writer {
	if queueSize <= 0 {
		queueSize = 1024
	}
	if duplicateWindow <= 0 {
		duplicateWindow = 30 * time.Second
	}
	return &Writer{navigator: navigator, log: log, queue: make(chan roomentered.Payload, queueSize), duplicateWindow: duplicateWindow, now: time.Now, batchSize: 64}
}

// Register subscribes room admission without blocking its owner loop.
func Register(lifecycle fx.Lifecycle, subscriber bus.Subscriber, writer *Writer) error {
	ctx, cancel := context.WithCancel(context.Background())
	writer.cancel = cancel
	writer.wait.Add(1)
	go writer.run(ctx)
	subscription, err := subscriber.Subscribe(roomentered.Name, bus.PriorityLow, func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(roomentered.Payload)
		if ok && payload.PlayerID > 0 && payload.RoomID > 0 {
			writer.Enqueue(payload)
		}
		return nil
	})
	if err != nil {
		cancel()
		writer.wait.Wait()
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		subscription.Unsubscribe()
		cancel()
		writer.wait.Wait()
		return nil
	}})
	return nil
}

// Enqueue submits one visit without allocating a goroutine or blocking admission.
func (writer *Writer) Enqueue(payload roomentered.Payload) bool {
	select {
	case writer.queue <- payload:
		return true
	default:
		writer.dropped.Add(1)
		return false
	}
}

// Dropped returns the number of visits rejected by bounded backpressure.
func (writer *Writer) Dropped() uint64 { return writer.dropped.Load() }

// run persists queued visits until shutdown.
func (writer *Writer) run(ctx context.Context) {
	defer writer.wait.Done()
	last := make(map[visitKey]time.Time)
	for {
		select {
		case <-ctx.Done():
			writer.drain(last)
			return
		case payload := <-writer.queue:
			writer.persistBatch(ctx, writer.accepted(payload), last)
		}
	}
}

// visitKey identifies a player-room pair in the short deduplication window.
type visitKey struct {
	// playerID identifies the visitor.
	playerID int64
	// roomID identifies the visited room.
	roomID int64
}

// accepted drains one bounded group including the first accepted event.
func (writer *Writer) accepted(first roomentered.Payload) []roomentered.Payload {
	values := make([]roomentered.Payload, 1, writer.batchSize)
	values[0] = first
	for len(values) < writer.batchSize {
		select {
		case payload := <-writer.queue:
			values = append(values, payload)
		default:
			return values
		}
	}
	return values
}

// persistBatch coalesces player-room pairs and performs one persistence call.
func (writer *Writer) persistBatch(ctx context.Context, payloads []roomentered.Payload, last map[visitKey]time.Time) {
	visits := make([]navrecord.Visit, 0, len(payloads))
	for _, payload := range payloads {
		key := visitKey{playerID: payload.PlayerID, roomID: payload.RoomID}
		now := writer.now()
		increment := true
		if previous, found := last[key]; found && now.Sub(previous) < writer.duplicateWindow {
			increment = false
		}
		last[key] = now
		merged := false
		for index := range visits {
			if visits[index].PlayerID == payload.PlayerID && visits[index].RoomID == payload.RoomID {
				visits[index].VisitedAt = now
				visits[index].Increment = visits[index].Increment || increment
				merged = true
				break
			}
		}
		if !merged {
			visits = append(visits, navrecord.Visit{PlayerID: payload.PlayerID, RoomID: payload.RoomID, VisitedAt: now, Increment: increment})
		}
	}
	recorder, ok := writer.navigator.(interface {
		RecordVisits(context.Context, []navrecord.Visit) error
	})
	if !ok {
		for _, visit := range visits {
			if visit.Increment {
				_ = writer.navigator.RecordVisit(ctx, visit.PlayerID, visit.RoomID)
			}
		}
		return
	}
	if err := recorder.RecordVisits(ctx, visits); err != nil && writer.log != nil {
		writer.log.Warn("record navigator room visits", zap.Int("count", len(visits)), zap.Error(err))
	}
}

// drain flushes already accepted visits without extending shutdown indefinitely.
func (writer *Writer) drain(last map[visitKey]time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for {
		select {
		case payload := <-writer.queue:
			writer.persistBatch(ctx, writer.accepted(payload), last)
		default:
			return
		}
	}
}
