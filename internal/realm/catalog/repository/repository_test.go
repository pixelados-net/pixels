package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// fakeExecutor records catalog repository calls.
type fakeExecutor struct {
	// row stores the next row result.
	row pgx.Row

	// rows stores the next rows result.
	rows pgx.Rows

	// tag stores the next command result.
	tag pgconn.CommandTag

	// query stores the last statement.
	query string

	// arguments stores the last statement arguments.
	arguments []any
}

// Exec records one command.
func (executor *fakeExecutor) Exec(_ context.Context, query string, arguments ...any) (pgconn.CommandTag, error) {
	executor.query = query
	executor.arguments = arguments

	return executor.tag, nil
}

// Query records one rows query.
func (executor *fakeExecutor) Query(_ context.Context, query string, arguments ...any) (pgx.Rows, error) {
	executor.query = query
	executor.arguments = arguments

	return executor.rows, nil
}

// QueryRow records one row query.
func (executor *fakeExecutor) QueryRow(_ context.Context, query string, arguments ...any) pgx.Row {
	executor.query = query
	executor.arguments = arguments

	return executor.row
}

// fakeRow copies fixed values into scan destinations.
type fakeRow struct {
	// values stores scan source values.
	values []any

	// err stores an optional scan failure.
	err error
}

// Scan copies fixed row values.
func (row fakeRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	assignValues(destinations, row.values)

	return nil
}

// fakeRows iterates fixed row values.
type fakeRows struct {
	// values stores rows in scan order.
	values [][]any

	// index stores the next row index.
	index int

	// err stores an optional iteration failure.
	err error
}

// Close closes fake rows.
func (rows *fakeRows) Close() {}

// Err returns an iteration failure.
func (rows *fakeRows) Err() error { return rows.err }

// CommandTag returns an empty command tag.
func (rows *fakeRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

// FieldDescriptions returns no field metadata.
func (rows *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

// Next advances to the next row.
func (rows *fakeRows) Next() bool {
	if rows.index >= len(rows.values) {
		return false
	}
	rows.index++

	return true
}

// Scan copies the current row.
func (rows *fakeRows) Scan(destinations ...any) error {
	assignValues(destinations, rows.values[rows.index-1])

	return nil
}

// Values returns the current row values.
func (rows *fakeRows) Values() ([]any, error) { return rows.values[rows.index-1], nil }

// RawValues returns no raw values.
func (rows *fakeRows) RawValues() [][]byte { return nil }

// Conn returns no backing connection.
func (rows *fakeRows) Conn() *pgx.Conn { return nil }

// assignValues copies test values into scan destinations.
func assignValues(destinations []any, values []any) {
	for index, value := range values {
		switch target := destinations[index].(type) {
		case *int:
			*target = value.(int)
		case *int32:
			*target = value.(int32)
		case *int64:
			*target = value.(int64)
		case *float64:
			*target = value.(float64)
		case *string:
			*target = value.(string)
		case *bool:
			*target = value.(bool)
		case *[]byte:
			*target = value.([]byte)
		case *time.Time:
			*target = value.(time.Time)
		case *pgtype.Int8:
			*target = value.(pgtype.Int8)
		case *pgtype.Timestamptz:
			*target = value.(pgtype.Timestamptz)
		}
	}
}

// pageValuesForTest returns one scannable page row.
func pageValuesForTest() []any {
	now := time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)
	return []any{int64(1), pgtype.Int8{}, "chairs", "default_3x3", int32(1), int32(2), pgtype.Text{}, int32(0),
		true, true, false, false, pgtype.Timestamptz{}, false, now, now, pgtype.Timestamptz{}, int64(1)}
}

// itemValuesForTest returns one scannable offer row.
func itemValuesForTest() []any {
	now := time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)
	return []any{int64(2), int64(1), pgtype.Int8{Int64: 3, Valid: true}, pgtype.Int8{}, pgtype.Int4{}, int32(0), "chair_plasto", int64(10), int64(0), int32(-1), int32(1), int32(0), int32(0),
		false, false, false, int32(0), true, "0", pgtype.Timestamptz{}, now, now, pgtype.Timestamptz{}, int64(1)}
}

// TestRepositoryTransactionRunsWork verifies the injected transaction boundary.
func TestRepositoryTransactionRunsWork(t *testing.T) {
	repository := newRepository(&fakeExecutor{})
	called := false
	err := repository.WithinTransaction(context.Background(), func(context.Context) error {
		called = true
		return nil
	})
	if err != nil || !called {
		t.Fatalf("expected transaction work, called=%v err=%v", called, err)
	}
	if New(nil) == nil {
		t.Fatal("expected pool repository")
	}
}
