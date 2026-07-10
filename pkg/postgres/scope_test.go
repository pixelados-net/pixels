package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestScopeSelectsExecutorAndDefersCallbacks verifies transaction context behavior.
func TestScopeSelectsExecutorAndDefersCallbacks(t *testing.T) {
	active := &scopeExecutor{}
	fallback := &scopeExecutor{}
	transactionScope := &scope{executor: active}
	ctx := context.WithValue(context.Background(), scopeKey{}, transactionScope)

	if ExecutorFor(ctx, fallback) != active {
		t.Fatal("expected scoped executor")
	}
	called := false
	if !AfterCommit(ctx, func(context.Context) { called = true }) {
		t.Fatal("expected callback registration")
	}
	if called {
		t.Fatal("expected deferred callback")
	}
	transactionScope.run(context.Background())
	if !called {
		t.Fatal("expected callback after commit")
	}
}

// TestWithinScopeWrapsBeginFailure verifies transaction startup failures.
func TestWithinScopeWrapsBeginFailure(t *testing.T) {
	config := testConfig()
	config.Port = 1
	config.ConnectTimeout = time.Millisecond
	poolConfig, err := pgxpool.ParseConfig(config.DSN())
	if err != nil {
		t.Fatalf("parse pool config: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	t.Cleanup(pool.Close)

	err = WithinScope(context.Background(), pool, func(context.Context) error {
		t.Fatal("unexpected scope callback")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "begin postgres transaction") {
		t.Fatalf("expected wrapped begin error, got %v", err)
	}
}

// TestScopeFallsBackOutsideTransaction verifies ordinary repository execution.
func TestScopeFallsBackOutsideTransaction(t *testing.T) {
	fallback := &scopeExecutor{}
	if ExecutorFor(context.Background(), fallback) != fallback {
		t.Fatal("expected fallback executor")
	}
	if AfterCommit(context.Background(), func(context.Context) {}) {
		t.Fatal("expected immediate publication outside transaction")
	}
}

// scopeExecutor implements Executor for transaction scope tests.
type scopeExecutor struct{}

// Exec accepts one test command.
func (*scopeExecutor) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

// Query accepts one test rows query.
func (*scopeExecutor) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

// QueryRow accepts one test row query.
func (*scopeExecutor) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}
