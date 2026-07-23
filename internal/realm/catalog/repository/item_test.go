package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestItemPersistenceScansAndWrites verifies offer list and creation.
func TestItemPersistenceScansAndWrites(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{itemValuesForTest()}}}
	items, err := newRepository(executor).ListItems(context.Background(), nil)
	if err != nil || len(items) != 1 || items[0].Name != "chair_plasto" {
		t.Fatalf("unexpected items=%#v err=%v", items, err)
	}
	if len(executor.arguments) != 1 || !strings.Contains(executor.query, "$1::bigint is null") {
		t.Fatalf("unexpected list query=%q args=%#v", executor.query, executor.arguments)
	}

	executor.row = fakeRow{values: itemValuesForTest()}
	created, err := newRepository(executor).CreateItem(context.Background(), catalogmodel.Item{PageID: 1, DefinitionID: 3, Name: "chair_plasto", Amount: 1, Enabled: true})
	if err != nil || created.ID != 2 || len(executor.arguments) != 22 {
		t.Fatalf("unexpected created=%#v args=%#v err=%v", created, executor.arguments, err)
	}
}

// TestFindItemByIDReturnsOffer verifies direct offer lookup.
func TestFindItemByIDReturnsOffer(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: itemValuesForTest()}}
	item, found, err := newRepository(executor).FindItemByID(context.Background(), 2)
	if err != nil || !found || item.ID != 2 {
		t.Fatalf("unexpected item=%#v found=%v err=%v", item, found, err)
	}
}

// TestItemUpdatesAndSoftDeleteUseVersions verifies optimistic offer mutations.
func TestItemUpdatesAndSoftDeleteUseVersions(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}
	item := catalogmodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}, Version: sharedmodel.Version{Version: 3}}}
	_, updated, err := newRepository(executor).UpdateItem(context.Background(), item)
	if err != nil || updated || len(executor.arguments) != 24 {
		t.Fatalf("unexpected update=%v args=%#v err=%v", updated, executor.arguments, err)
	}

	executor.tag = pgconn.NewCommandTag("UPDATE 1")
	deleted, err := newRepository(executor).SoftDeleteItem(context.Background(), 2, 3)
	if err != nil || !deleted || len(executor.arguments) != 2 {
		t.Fatalf("unexpected delete=%v args=%#v err=%v", deleted, executor.arguments, err)
	}
}

// TestSanitizeListReturnsOnlyDefinitionsWithoutOffer verifies sanitation projection.
func TestSanitizeListReturnsOnlyDefinitionsWithoutOffer(t *testing.T) {
	values := definitionValuesForTest()
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{values}}}
	definitions, err := newRepository(executor).SanitizeList(context.Background())
	if err != nil || len(definitions) != 1 || definitions[0].ID != 3 {
		t.Fatalf("unexpected definitions=%#v err=%v", definitions, err)
	}
	if !strings.Contains(executor.query, "i.id is null") {
		t.Fatalf("unexpected sanitize query %q", executor.query)
	}

	executor.row = fakeRow{values: []any{int64(1)}}
	count, err := newRepository(executor).CountEnabledDefinitionsWithoutOffer(context.Background())
	if err != nil || count != 1 {
		t.Fatalf("unexpected count=%d err=%v", count, err)
	}
}

// definitionValuesForTest returns one orphan definition row.
func definitionValuesForTest() []any {
	values := []any{
		int64(3), 39, "orphan_chair", "Orphan Chair", "floor", 1, 1, 1.0,
		false, false, true, false, true, "default", 2, "", "", []byte(`{}`),
	}
	values = append(values, itemValuesForTest()[23], itemValuesForTest()[24], pgtype.Timestamptz{}, int64(1))

	return values
}
