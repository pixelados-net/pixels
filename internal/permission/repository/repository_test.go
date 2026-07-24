package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// fakeExecutor records permission repository operations.
type fakeExecutor struct {
	// row stores the next row result.
	row pgx.Row
	// rows stores the next rows result.
	rows pgx.Rows
	// query stores the last SQL statement.
	query string
	// arguments stores the last SQL arguments.
	arguments []any
	// err stores the next executor failure.
	err error
}

// Exec records one statement.
func (executor *fakeExecutor) Exec(_ context.Context, query string, arguments ...any) (pgconn.CommandTag, error) {
	executor.query = query
	executor.arguments = arguments
	return pgconn.CommandTag{}, executor.err
}

// Query records one collection query.
func (executor *fakeExecutor) Query(_ context.Context, query string, arguments ...any) (pgx.Rows, error) {
	executor.query = query
	executor.arguments = arguments
	return executor.rows, executor.err
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
		case *int32:
			*target = value.(int32)
		case *int64:
			*target = value.(int64)
		case *string:
			*target = value.(string)
		case *bool:
			*target = value.(bool)
		case *time.Time:
			*target = value.(time.Time)
		case *pgtype.Int8:
			*target = value.(pgtype.Int8)
		case *pgtype.Int4:
			*target = value.(pgtype.Int4)
		case *pgtype.Timestamptz:
			*target = value.(pgtype.Timestamptz)
		}
	}
}

// groupValues returns one scannable permission group row.
func groupValues() []any {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	return []any{int64(2), "admin", int32(100), "ADMIN", "#ff0000", "https://cdn.example/admin.png",
		pgtype.Int4{Int32: 42, Valid: true}, pgtype.Int8{Int64: 1, Valid: true}, now, now, pgtype.Timestamptz{}, int64(3)}
}
