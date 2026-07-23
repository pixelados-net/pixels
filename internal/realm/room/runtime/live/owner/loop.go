// Package owner runs the single periodic owner loop for an active room.
package owner

import (
	"context"
	"sync"
	"time"
)

// Tick performs one room owner cycle.
type Tick func(context.Context)

// Loop owns one cancellable periodic room task.
type Loop struct {
	// mutex protects loop lifecycle state.
	mutex sync.Mutex
	// cancel stops the active loop.
	cancel context.CancelFunc
	// done closes after the active loop returns.
	done chan struct{}
}

// Start starts the loop once when interval and tick are valid.
func (loop *Loop) Start(ctx context.Context, interval time.Duration, tick Tick) {
	if interval <= 0 || tick == nil {
		return
	}
	loop.mutex.Lock()
	if loop.cancel != nil {
		loop.mutex.Unlock()
		return
	}
	loopCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	loop.cancel = cancel
	loop.done = done
	loop.mutex.Unlock()
	go run(loopCtx, interval, tick, done)
}

// Stop stops the active loop and waits for its completion.
func (loop *Loop) Stop() {
	loop.mutex.Lock()
	cancel := loop.cancel
	done := loop.done
	loop.cancel = nil
	loop.done = nil
	loop.mutex.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	<-done
}

// run invokes a task on every interval until cancellation.
func run(ctx context.Context, interval time.Duration, tick Tick, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick(ctx)
		}
	}
}
