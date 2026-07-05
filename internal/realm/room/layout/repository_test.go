package layout

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// TestRepositoryCreateScansLayout verifies layout creation scans returned fields.
func TestRepositoryCreateScansLayout(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: layoutValuesForTest()}}

	roomLayout, err := NewRepository(executor).Create(context.Background(), CreateRecordParams{Layout: validLayoutForTest()})
	if err != nil {
		t.Fatalf("create layout: %v", err)
	}

	if roomLayout.ID != 7 || roomLayout.Name != "model_a" {
		t.Fatalf("unexpected layout %#v", roomLayout)
	}

	if !strings.Contains(executor.query, "insert into room_layouts") {
		t.Fatalf("expected insert query, got %q", executor.query)
	}
}

// TestRepositoryUpdateReportsMissingLayout verifies missing updates are reported.
func TestRepositoryUpdateReportsMissingLayout(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}

	_, found, err := NewRepository(executor).Update(context.Background(), UpdateRecordParams{ID: 77, Layout: validLayoutForTest()})
	if err != nil {
		t.Fatalf("update layout: %v", err)
	}

	if found {
		t.Fatal("expected missing layout")
	}
}

// TestRepositoryFindByNameReportsMissingLayout verifies missing names are reported.
func TestRepositoryFindByNameReportsMissingLayout(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}

	_, found, err := NewRepository(executor).FindByName(context.Background(), "model_missing")
	if err != nil {
		t.Fatalf("find layout: %v", err)
	}

	if found {
		t.Fatal("expected missing layout")
	}
}

// TestRepositoryListScansLayouts verifies list scans all rows.
func TestRepositoryListScansLayouts(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{layoutValuesForTest()}}}

	layouts, err := NewRepository(executor).List(context.Background())
	if err != nil {
		t.Fatalf("list layouts: %v", err)
	}

	if len(layouts) != 1 || layouts[0].Name != "model_a" {
		t.Fatalf("unexpected layouts %#v", layouts)
	}
}

// TestRepositoryListWrapsRowsError verifies row iteration failures are wrapped.
func TestRepositoryListWrapsRowsError(t *testing.T) {
	expected := errors.New("rows failed")
	executor := &fakeExecutor{rows: &fakeRows{err: expected}}

	_, err := NewRepository(executor).List(context.Background())
	if !errors.Is(err, expected) {
		t.Fatalf("expected rows error, got %v", err)
	}
}

// fakeExecutor records repository query calls for tests.
type fakeExecutor struct {
	// row is the row returned by QueryRow.
	row pgx.Row

	// rows are the rows returned by Query.
	rows pgx.Rows

	// query is the last executed query.
	query string
}

// Exec executes SQL without returning rows.
func (executor *fakeExecutor) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

// Query executes SQL returning multiple rows.
func (executor *fakeExecutor) Query(_ context.Context, query string, _ ...any) (pgx.Rows, error) {
	executor.query = query

	return executor.rows, nil
}

// QueryRow executes SQL returning one row.
func (executor *fakeExecutor) QueryRow(_ context.Context, query string, _ ...any) pgx.Row {
	executor.query = query

	return executor.row
}

// fakeRow scans fixed values for tests.
type fakeRow struct {
	// values are copied into scan destinations.
	values []any

	// err is returned by Scan before copying values.
	err error
}

// Scan copies values into destinations.
func (row fakeRow) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}

	assign(destinations, row.values)

	return nil
}

// fakeRows scans fixed row values for tests.
type fakeRows struct {
	// values are copied into scan destinations per row.
	values [][]any

	// index stores the next row index.
	index int

	// closed reports whether rows were closed.
	closed bool

	// err is returned by Err.
	err error
}

// Close closes the rows.
func (rows *fakeRows) Close() {
	rows.closed = true
}

// Err returns the rows error.
func (rows *fakeRows) Err() error {
	return rows.err
}

// CommandTag returns the rows command tag.
func (rows *fakeRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

// FieldDescriptions returns the rows field descriptions.
func (rows *fakeRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

// Next advances to the next row.
func (rows *fakeRows) Next() bool {
	if rows.index >= len(rows.values) {
		rows.Close()
		return false
	}

	rows.index++

	return true
}

// Scan copies the current row into destinations.
func (rows *fakeRows) Scan(destinations ...any) error {
	assign(destinations, rows.values[rows.index-1])

	return nil
}

// Values returns the current row values.
func (rows *fakeRows) Values() ([]any, error) {
	return rows.values[rows.index-1], nil
}

// RawValues returns raw values.
func (rows *fakeRows) RawValues() [][]byte {
	return nil
}

// Conn returns the rows connection.
func (rows *fakeRows) Conn() *pgx.Conn {
	return nil
}

// assign copies values into destinations.
func assign(destinations []any, values []any) {
	for index, value := range values {
		assignValue(destinations[index], value)
	}
}

// assignValue copies one value into one destination.
func assignValue(destination any, value any) {
	switch target := destination.(type) {
	case *int:
		*target = value.(int)
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

// validLayoutForTest returns valid editable layout values.
func validLayoutForTest() Layout {
	return Layout{Name: "model_a", TileSize: 12, Heightmap: "xxx\rx0x\rxxx", DoorX: 1, DoorY: 1, DoorDirection: 2, Enabled: true}
}

// layoutValuesForTest returns a scannable room layout row.
func layoutValuesForTest() []any {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)

	return []any{int64(7), "model_a", 12, "xxx\rx0x\rxxx", 1, 1, 0, 2, 0, true, now, now, pgtype.Timestamptz{}, int64(1)}
}
