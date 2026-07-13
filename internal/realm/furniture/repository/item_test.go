package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

// TestCreateItemsInsertsAllRowsInOneStatement verifies bulk inventory creation.
func TestCreateItemsInsertsAllRowsInOneStatement(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{
		itemValuesForTest(pgtype.Int8{}, pgtype.Int2{}, pgtype.Int2{}, pgtype.Float8{}),
		itemValuesForTest(pgtype.Int8{}, pgtype.Int2{}, pgtype.Int2{}, pgtype.Float8{}),
	}}}

	items, err := New(executor).CreateItems(context.Background(), 2, 7, 2, "0", nil)
	if err != nil {
		t.Fatalf("create items: %v", err)
	}
	if len(items) != 2 || !items[0].InInventory() || !items[1].InInventory() {
		t.Fatalf("unexpected created items %#v", items)
	}
	if !strings.Contains(executor.query, "generate_series") || len(executor.arguments) != 5 {
		t.Fatalf("expected one bulk statement, query=%q arguments=%#v", executor.query, executor.arguments)
	}
}
