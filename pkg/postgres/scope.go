package postgres

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
)

// scopeKey identifies a transaction scope in context.
type scopeKey struct{}

// scope contains one active executor and deferred commit callbacks.
type scope struct {
	// executor runs statements in the active transaction.
	executor Executor

	// mutex protects callback registration.
	mutex sync.Mutex

	// callbacks run only after a successful commit.
	callbacks []func(context.Context)
}

// WithinScope runs work in one transaction shared through context.
func WithinScope(ctx context.Context, pool *Pool, work func(context.Context) error) error {
	transactionScope := &scope{}
	err := WithinTx(ctx, pool, func(_ context.Context, tx pgx.Tx) error {
		transactionScope.executor = tx
		return work(context.WithValue(ctx, scopeKey{}, transactionScope))
	})
	if err != nil {
		return err
	}

	transactionScope.run(ctx)

	return nil
}

// ScopedExecutor returns the active transaction executor when present.
func ScopedExecutor(ctx context.Context) (Executor, bool) {
	transactionScope, ok := ctx.Value(scopeKey{}).(*scope)
	if !ok || transactionScope.executor == nil {
		return nil, false
	}

	return transactionScope.executor, true
}

// ExecutorFor returns the active transaction executor or a fallback.
func ExecutorFor(ctx context.Context, fallback Executor) Executor {
	executor, ok := ScopedExecutor(ctx)
	if ok {
		return executor
	}

	return fallback
}

// AfterCommit defers a callback when a transaction scope is active.
func AfterCommit(ctx context.Context, callback func(context.Context)) bool {
	transactionScope, ok := ctx.Value(scopeKey{}).(*scope)
	if !ok {
		return false
	}

	transactionScope.mutex.Lock()
	transactionScope.callbacks = append(transactionScope.callbacks, callback)
	transactionScope.mutex.Unlock()

	return true
}

// run executes deferred callbacks in registration order.
func (transactionScope *scope) run(ctx context.Context) {
	transactionScope.mutex.Lock()
	callbacks := append([]func(context.Context){}, transactionScope.callbacks...)
	transactionScope.mutex.Unlock()

	for _, callback := range callbacks {
		callback(ctx)
	}
}
