package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
)

// TestRepositoryFindCustomByRoomIDScansLayout verifies custom layout reads.
func TestRepositoryFindCustomByRoomIDScansLayout(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: customValuesForTest()}}
	roomLayout, found, err := NewRepository(executor).FindCustomByRoomID(context.Background(), 9)
	if err != nil {
		t.Fatalf("find custom layout: %v", err)
	}
	if !found || roomLayout.RoomID != 9 || roomLayout.WallHeight != -1 {
		t.Fatalf("unexpected custom layout %#v found=%v", roomLayout, found)
	}
}

// TestRepositoryFindCustomByRoomIDReportsMissing verifies absent custom layouts.
func TestRepositoryFindCustomByRoomIDReportsMissing(t *testing.T) {
	_, found, err := NewRepository(&fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}).FindCustomByRoomID(context.Background(), 9)
	if err != nil || found {
		t.Fatalf("expected missing custom layout found=%v err=%v", found, err)
	}
}

// TestRepositoryUpsertCustomSynchronizesRoom verifies custom save query behavior.
func TestRepositoryUpsertCustomSynchronizesRoom(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: customValuesForTest()}}
	repository := NewRepository(executor)
	roomLayout, err := repository.UpsertCustom(context.Background(), roomlayout.CustomSaveParams{
		RoomID: 9, Heightmap: "00", DoorDirection: 2, WallHeight: -1,
	})
	if err != nil {
		t.Fatalf("upsert custom layout: %v", err)
	}
	if roomLayout.RoomID != 9 || executor.query != upsertCustomSQL {
		t.Fatalf("unexpected custom save %#v query=%q", roomLayout, executor.query)
	}
	called := false
	if err = repository.WithinTransaction(context.Background(), func(context.Context) error { called = true; return nil }); err != nil || !called {
		t.Fatalf("run fallback transaction called=%v err=%v", called, err)
	}
}

// TestRepositoryFixedOperationsCoverSuccessfulReadsAndUpdates verifies ordinary repository paths.
func TestRepositoryFixedOperationsCoverSuccessfulReadsAndUpdates(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: layoutValuesForTest()}}
	repository := NewRepository(executor)
	if _, found, err := repository.FindByID(context.Background(), 7); err != nil || !found {
		t.Fatalf("find layout found=%v err=%v", found, err)
	}
	if _, found, err := repository.Update(context.Background(), roomlayout.UpdateRecordParams{ID: 7, Layout: validLayoutForTest()}); err != nil || !found {
		t.Fatalf("update layout found=%v err=%v", found, err)
	}
}

// customValuesForTest returns one scannable custom layout row.
func customValuesForTest() []any {
	return []any{int64(9), "model_a", "00", 0, 0, 2, -1, 1, -1, time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)}
}
