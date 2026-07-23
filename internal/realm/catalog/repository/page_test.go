package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestPagePersistenceScansAndWrites verifies page read and create statements.
func TestPagePersistenceScansAndWrites(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{pageValuesForTest()}}}
	pages, err := newRepository(executor).ListPages(context.Background())
	if err != nil || len(pages) != 1 || pages[0].Name != "chairs" {
		t.Fatalf("unexpected pages=%#v err=%v", pages, err)
	}

	executor.row = fakeRow{values: pageValuesForTest()}
	created, err := newRepository(executor).CreatePage(context.Background(), catalogmodel.Page{Name: "chairs", Layout: catalogmodel.DefaultLayout, Visible: true, Enabled: true})
	if err != nil || created.ID != 1 || len(executor.arguments) != 13 {
		t.Fatalf("unexpected created=%#v args=%#v err=%v", created, executor.arguments, err)
	}
}

// TestFindAndUpdatePageHandleOptimisticResults verifies optional page writes.
func TestFindAndUpdatePageHandleOptimisticResults(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: pageValuesForTest()}}
	page, found, err := newRepository(executor).FindPageByID(context.Background(), 1)
	if err != nil || !found || page.ID != 1 {
		t.Fatalf("unexpected page=%#v found=%v err=%v", page, found, err)
	}

	executor.row = fakeRow{err: pgx.ErrNoRows}
	_, updated, err := newRepository(executor).UpdatePage(context.Background(), catalogmodel.Page{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}, Version: sharedmodel.Version{Version: 2}}})
	if err != nil || updated {
		t.Fatalf("expected optimistic miss, updated=%v err=%v", updated, err)
	}
}
