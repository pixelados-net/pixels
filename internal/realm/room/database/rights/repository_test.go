package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// executorForTest stores deterministic PostgreSQL results.
type executorForTest struct {
	// tag stores mutation results.
	tag pgconn.CommandTag
	// exists stores the QueryRow result.
	exists bool
	// rows stores query rights.
	rows *rowsForTest
	// err optionally fails executor operations.
	err error
}

// Exec returns a configured command tag.
func (executor *executorForTest) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return executor.tag, executor.err
}

// Query returns configured rights rows.
func (executor *executorForTest) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return executor.rows, executor.err
}

// QueryRow returns one exists row.
func (executor *executorForTest) QueryRow(context.Context, string, ...any) pgx.Row {
	return rowForTest{exists: executor.exists, err: executor.err}
}

// rowForTest scans an exists result.
type rowForTest struct {
	// exists stores the scanned decision.
	exists bool
	// err optionally fails scanning.
	err error
}

// Scan copies the exists result.
func (row rowForTest) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	*destinations[0].(*bool) = row.exists
	return nil
}

// TestRepositoryWrapsExecutorErrors verifies contextual persistence errors.
func TestRepositoryWrapsExecutorErrors(t *testing.T) {
	expected := errors.New("database unavailable")
	repository := &Repository{executor: &executorForTest{err: expected}}
	ctx := context.Background()
	if _, err := repository.Grant(ctx, 9, 2, 1); !errors.Is(err, expected) {
		t.Fatalf("expected grant error, got %v", err)
	}
	if _, err := repository.List(ctx, 9); !errors.Is(err, expected) {
		t.Fatalf("expected list error, got %v", err)
	}
	if _, err := repository.Exists(ctx, 9, 2); !errors.Is(err, expected) {
		t.Fatalf("expected exists error, got %v", err)
	}
}

// rowsForTest scans room rights rows.
type rowsForTest struct {
	// remaining stores unread rows.
	remaining int
	// current stores whether Next advanced.
	current bool
}

// Close closes rows.
func (*rowsForTest) Close() {}

// Err returns no row error.
func (*rowsForTest) Err() error { return nil }

// CommandTag returns an empty tag.
func (*rowsForTest) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

// FieldDescriptions returns no descriptions.
func (*rowsForTest) FieldDescriptions() []pgconn.FieldDescription { return nil }

// Next advances one row.
func (rows *rowsForTest) Next() bool {
	if rows.remaining == 0 {
		return false
	}
	rows.remaining--
	rows.current = true
	return true
}

// Scan writes one rights row.
func (rows *rowsForTest) Scan(destinations ...any) error {
	*destinations[0].(*int64) = 9
	*destinations[1].(*int64) = 2
	*destinations[2].(*string) = "Alice"
	*destinations[3].(*int64) = 1
	*destinations[4].(*time.Time) = time.Unix(1000, 0)
	return nil
}

// Values returns no raw values.
func (*rowsForTest) Values() ([]any, error) { return nil, nil }

// RawValues returns no raw values.
func (*rowsForTest) RawValues() [][]byte { return nil }

// Conn returns no physical connection.
func (*rowsForTest) Conn() *pgx.Conn { return nil }

// TestRepositoryMutatesAndReadsRights verifies all indexed repository paths.
func TestRepositoryMutatesAndReadsRights(t *testing.T) {
	executor := &executorForTest{tag: pgconn.NewCommandTag("INSERT 0 1"), exists: true, rows: &rowsForTest{remaining: 1}}
	repository := &Repository{executor: executor, withinTx: func(ctx context.Context, work func(context.Context) error) error { return work(ctx) }}
	ctx := context.Background()
	if created, err := repository.Grant(ctx, 9, 2, 1); err != nil || !created {
		t.Fatalf("grant created=%v err=%v", created, err)
	}
	if removed, err := repository.Revoke(ctx, 9, 2); err != nil || !removed {
		t.Fatalf("revoke removed=%v err=%v", removed, err)
	}
	if exists, err := repository.Exists(ctx, 9, 2); err != nil || !exists {
		t.Fatalf("exists=%v err=%v", exists, err)
	}
	rights, err := repository.List(ctx, 9)
	if err != nil || len(rights) != 1 || rights[0].Username != "Alice" {
		t.Fatalf("rights=%#v err=%v", rights, err)
	}
}

// TestRepositoryRevokeAllAndTransaction verifies returning rows and transaction work.
func TestRepositoryRevokeAllAndTransaction(t *testing.T) {
	repository := &Repository{executor: &executorForTest{rows: &rowsForTest{remaining: 1}}, withinTx: func(ctx context.Context, work func(context.Context) error) error { return work(ctx) }}
	rights, err := repository.RevokeAll(context.Background(), 9)
	if err != nil || len(rights) != 1 || rights[0].PlayerID != 2 {
		t.Fatalf("rights=%#v err=%v", rights, err)
	}
	called := false
	err = repository.WithinTransaction(context.Background(), func(context.Context) error { called = true; return nil })
	if err != nil || !called {
		t.Fatalf("transaction called=%v err=%v", called, err)
	}
}
