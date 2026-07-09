package repository

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeRow scans fixed values.
type fakeRow struct {
	// values stores scan values.
	values []any
	// err stores a scan error.
	err error
}

// Scan copies fixed values into destinations.
func (row fakeRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	for index, value := range row.values {
		switch target := destinations[index].(type) {
		case *int32:
			*target = value.(int32)
		case *int64:
			*target = value.(int64)
		case *time.Time:
			*target = value.(time.Time)
		}
	}
	return nil
}

// fakeRows iterates fixed result rows.
type fakeRows struct {
	// values stores row values.
	values [][]any
	// index stores the current row.
	index int
	// err stores an iteration error.
	err error
}

// Close closes fake rows.
func (rows *fakeRows) Close() {}

// Err returns the iteration error.
func (rows *fakeRows) Err() error { return rows.err }

// CommandTag returns an empty command tag.
func (rows *fakeRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

// FieldDescriptions returns no field descriptions.
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
	return fakeRow{values: rows.values[rows.index-1]}.Scan(destinations...)
}

// Values returns the current row values.
func (rows *fakeRows) Values() ([]any, error) {
	if rows.index == 0 {
		return nil, errors.New("no current row")
	}
	return rows.values[rows.index-1], nil
}

// RawValues returns no raw values.
func (rows *fakeRows) RawValues() [][]byte { return nil }

// Conn returns no concrete connection.
func (rows *fakeRows) Conn() *pgx.Conn { return nil }
