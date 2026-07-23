package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// fakeExecutor records repository query calls for tests.
type fakeExecutor struct {
	// row is returned by QueryRow.
	row pgx.Row

	// rows are returned by Query.
	rows pgx.Rows

	// tag is returned by Exec.
	tag pgconn.CommandTag

	// execs stores Exec call count.
	execs int
}

// Exec executes SQL without returning rows.
func (executor *fakeExecutor) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	executor.execs++
	return executor.tag, nil
}

// Query executes SQL returning rows.
func (executor *fakeExecutor) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return executor.rows, nil
}

// QueryRow executes SQL returning one row.
func (executor *fakeExecutor) QueryRow(context.Context, string, ...any) pgx.Row {
	return executor.row
}

// fakeRow scans fixed values for tests.
type fakeRow struct {
	// values are copied into destinations.
	values []any

	// err is returned before copying values.
	err error
}

// Scan copies values into destinations.
func (row fakeRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	assignValues(destinations, row.values)
	return nil
}

// fakeRows scans fixed row values for tests.
type fakeRows struct {
	// values are copied into destinations per row.
	values [][]any

	// index stores the next row index.
	index int

	// err is returned by Err.
	err error
}

// Close closes rows.
func (rows *fakeRows) Close() {}

// Err returns the rows error.
func (rows *fakeRows) Err() error { return rows.err }

// CommandTag returns the rows command tag.
func (rows *fakeRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

// FieldDescriptions returns field descriptions.
func (rows *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

// Next advances to the next row.
func (rows *fakeRows) Next() bool {
	if rows.index >= len(rows.values) {
		return false
	}
	rows.index++
	return true
}

// Scan copies current row values.
func (rows *fakeRows) Scan(destinations ...any) error {
	assignValues(destinations, rows.values[rows.index-1])
	return nil
}

// Values returns current row values.
func (rows *fakeRows) Values() ([]any, error) { return rows.values[rows.index-1], nil }

// RawValues returns raw values.
func (rows *fakeRows) RawValues() [][]byte { return nil }

// Conn returns the rows connection.
func (rows *fakeRows) Conn() *pgx.Conn { return nil }

// assignValues copies values into destinations.
func assignValues(destinations []any, values []any) {
	for index, value := range values {
		assignValue(destinations[index], value)
	}
}

// assignValue copies one value into one destination.
func assignValue(destination any, value any) {
	switch target := destination.(type) {
	case *int:
		*target = value.(int)
	case *int16:
		*target = value.(int16)
	case *int64:
		*target = value.(int64)
	case *string:
		*target = value.(string)
	case *bool:
		*target = value.(bool)
	case *time.Time:
		*target = value.(time.Time)
	case *pgtype.Timestamptz:
		*target = value.(pgtype.Timestamptz)
	}
}
