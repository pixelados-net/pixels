package votes

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeExecutor records vote repository operations.
type fakeExecutor struct {
	// tag stores the next execution result.
	tag pgconn.CommandTag
	// row stores the next single row.
	row pgx.Row
	// rows stores the next row set.
	rows pgx.Rows
	// execs stores execution count.
	execs int
	// queries stores query count.
	queries int
	// execErr stores execution failure.
	execErr error
	// queryErr stores query failure.
	queryErr error
}

// Exec returns the configured command tag.
func (executor *fakeExecutor) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	executor.execs++
	return executor.tag, executor.execErr
}

// Query returns configured rows.
func (executor *fakeExecutor) Query(context.Context, string, ...any) (pgx.Rows, error) {
	executor.queries++
	return executor.rows, executor.queryErr
}

// QueryRow returns the configured row.
func (executor *fakeExecutor) QueryRow(context.Context, string, ...any) pgx.Row { return executor.row }

// fakeRow scans one configured value set.
type fakeRow struct {
	// values stores scan source values.
	values []any
	// err stores a scan failure.
	err error
}

// Scan copies configured values.
func (row fakeRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	assign(destinations, row.values)
	return nil
}

// fakeRows scans configured vote rows.
type fakeRows struct {
	// values stores rows.
	values [][]any
	// index stores the next row.
	index int
	// err stores an iteration failure.
	err error
}

// Close closes rows.
func (*fakeRows) Close() {}

// Err reports no iteration error.
func (rows *fakeRows) Err() error { return rows.err }

// CommandTag returns an empty command tag.
func (*fakeRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

// FieldDescriptions returns no field metadata.
func (*fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

// Next advances one row.
func (rows *fakeRows) Next() bool {
	if rows.index >= len(rows.values) {
		return false
	}
	rows.index++
	return true
}

// Scan copies the current row.
func (rows *fakeRows) Scan(destinations ...any) error {
	assign(destinations, rows.values[rows.index-1])
	return nil
}

// Values returns the current row.
func (rows *fakeRows) Values() ([]any, error) { return rows.values[rows.index-1], nil }

// RawValues returns no raw values.
func (*fakeRows) RawValues() [][]byte { return nil }

// Conn returns no concrete connection.
func (*fakeRows) Conn() *pgx.Conn { return nil }

// assign copies supported test values.
func assign(destinations []any, values []any) {
	for index, value := range values {
		switch target := destinations[index].(type) {
		case *int:
			*target = value.(int)
		case *int64:
			*target = value.(int64)
		case *bool:
			*target = value.(bool)
		case *time.Time:
			*target = value.(time.Time)
		}
	}
}
