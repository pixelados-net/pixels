// Package work owns the bounded shared pet persistence workers.
package work

import (
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

// Pool executes deferred pet work without per-pet goroutines.
type Pool struct {
	// jobs stores bounded deferred work.
	jobs chan func()
	// stopped closes shared workers.
	stopped chan struct{}
	// workerCount stores the fixed concurrency.
	workerCount int
	// log records saturated queue warnings.
	log *zap.Logger
	// startOnce prevents duplicate worker starts.
	startOnce sync.Once
	// stopOnce prevents duplicate shutdown signals.
	stopOnce sync.Once
	// closed reports whether shutdown has begun.
	closed atomic.Bool
	// workers waits for shared workers to stop.
	workers sync.WaitGroup
}

// New creates one bounded fixed-size work pool.
func New(capacity int, workerCount int, log *zap.Logger) *Pool {
	if capacity <= 0 {
		capacity = 512
	}
	if workerCount <= 0 {
		workerCount = 2
	}
	return &Pool{jobs: make(chan func(), capacity), stopped: make(chan struct{}), workerCount: workerCount, log: log}
}

// Start launches the fixed worker set exactly once.
func (pool *Pool) Start() {
	pool.startOnce.Do(func() {
		pool.workers.Add(pool.workerCount)
		for range pool.workerCount {
			go pool.worker()
		}
	})
}

// Stop drains queued work and joins every shared worker.
func (pool *Pool) Stop() {
	pool.stopOnce.Do(func() {
		pool.closed.Store(true)
		close(pool.stopped)
	})
	pool.workers.Wait()
}

// Dispatch queues work and reports bounded acceptance.
func (pool *Pool) Dispatch(job func()) bool {
	if pool.closed.Load() {
		return false
	}
	select {
	case pool.jobs <- job:
		return true
	case <-pool.stopped:
		return false
	default:
		if pool.log != nil {
			pool.log.Warn("pet persistence queue full")
		}
		return false
	}
}

// worker executes queued work until shutdown has drained the queue.
func (pool *Pool) worker() {
	defer pool.workers.Done()
	for {
		select {
		case job := <-pool.jobs:
			if job != nil {
				job()
			}
		case <-pool.stopped:
			pool.drain()
			return
		}
	}
}

// drain executes the bounded queue remainder during shutdown.
func (pool *Pool) drain() {
	for {
		select {
		case job := <-pool.jobs:
			if job != nil {
				job()
			}
		default:
			return
		}
	}
}
