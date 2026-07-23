package database

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// TestCreateRequestReportsInsertion verifies idempotent request persistence.
func TestCreateRequestReportsInsertion(t *testing.T) {
	executor := &fakeExecutor{tag: pgconn.NewCommandTag("INSERT 0 1")}
	created, err := New(executor).CreateRequest(context.Background(), 1, 2)
	if err != nil || !created || len(executor.arguments) != 2 {
		t.Fatalf("unexpected created=%v args=%#v err=%v", created, executor.arguments, err)
	}
}

// TestCountFriendsScansCount verifies directional friend counting.
func TestCountFriendsScansCount(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{value: 3}}
	count, err := New(executor).CountFriends(context.Background(), 1)
	if err != nil || count != 3 {
		t.Fatalf("unexpected count=%d err=%v", count, err)
	}
}

// fakeExecutor records messenger persistence calls.
type fakeExecutor struct {
	// tag stores the next command result.
	tag pgconn.CommandTag
	// row stores the next row result.
	row pgx.Row
	// arguments stores the latest query arguments.
	arguments []any
}

// Exec records one command.
func (executor *fakeExecutor) Exec(_ context.Context, _ string, arguments ...any) (pgconn.CommandTag, error) {
	executor.arguments = arguments
	return executor.tag, nil
}

// Query supplies no row collection for focused tests.
func (*fakeExecutor) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }

// QueryRow returns one configured row.
func (executor *fakeExecutor) QueryRow(context.Context, string, ...any) pgx.Row { return executor.row }

// fakeRow scans one integer value.
type fakeRow struct {
	// value stores the count result.
	value int
}

// Scan copies the configured count.
func (row fakeRow) Scan(destinations ...any) error {
	*destinations[0].(*int) = row.value
	return nil
}
