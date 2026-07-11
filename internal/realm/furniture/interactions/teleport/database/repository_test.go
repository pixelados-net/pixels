package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	teleportpair "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport/pair"
)

// TestRepositoryFindReplaceAndDelete verifies teleport pair persistence behavior.
func TestRepositoryFindReplaceAndDelete(t *testing.T) {
	executor := &executorForTest{row: rowForTest{values: []int64{2, 9}}}
	repository := newRepository(executor)
	paired, found, err := repository.FindByItem(context.Background(), 2)
	if err != nil || !found || paired.ItemOneID != 2 || paired.ItemTwoID != 9 {
		t.Fatalf("unexpected pair %#v found=%v err=%v", paired, found, err)
	}
	if err := repository.Replace(context.Background(), paired); err != nil {
		t.Fatalf("replace pair: %v", err)
	}
	if len(executor.execSQL) != 2 {
		t.Fatalf("expected delete and insert, got %#v", executor.execSQL)
	}
	removed, err := repository.DeleteByItem(context.Background(), 2)
	if err != nil || !removed {
		t.Fatalf("delete pair removed=%v err=%v", removed, err)
	}
}

// TestRepositoryHandlesMissingAndStatementErrors verifies diagnostic branches.
func TestRepositoryHandlesMissingAndStatementErrors(t *testing.T) {
	repository := newRepository(&executorForTest{row: rowForTest{err: pgx.ErrNoRows}})
	if _, found, err := repository.FindByItem(context.Background(), 7); err != nil || found {
		t.Fatalf("expected missing pair found=%v err=%v", found, err)
	}
	expected := errors.New("database unavailable")
	repository = newRepository(&executorForTest{row: rowForTest{err: expected}, execErr: expected})
	if _, _, err := repository.FindByItem(context.Background(), 7); !errors.Is(err, expected) {
		t.Fatalf("expected find error, got %v", err)
	}
	if err := repository.Replace(context.Background(), teleportpair.Pair{ItemOneID: 1, ItemTwoID: 2}); !errors.Is(err, expected) {
		t.Fatalf("expected replace error, got %v", err)
	}
	if _, err := repository.DeleteByItem(context.Background(), 1); !errors.Is(err, expected) {
		t.Fatalf("expected delete error, got %v", err)
	}
}

// executorForTest records repository statements.
type executorForTest struct {
	// row stores the query row fixture.
	row rowForTest
	// execErr stores a statement failure.
	execErr error
	// execSQL stores executed statements.
	execSQL []string
}

// Exec records one statement.
func (executor *executorForTest) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	executor.execSQL = append(executor.execSQL, sql)
	if executor.execErr != nil {
		return pgconn.CommandTag{}, executor.execErr
	}

	return pgconn.NewCommandTag("DELETE 1"), nil
}

// Query is unused by the repository.
func (executor *executorForTest) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected query")
}

// QueryRow returns the configured row.
func (executor *executorForTest) QueryRow(context.Context, string, ...any) pgx.Row {
	return executor.row
}

// rowForTest scans fixed pair ids.
type rowForTest struct {
	// values stores item ids.
	values []int64
	// err stores a scan failure.
	err error
}

// Scan writes fixed pair ids.
func (row rowForTest) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	for index, value := range row.values {
		*destinations[index].(*int64) = value
	}

	return nil
}
