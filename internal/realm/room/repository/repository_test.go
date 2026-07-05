package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
)

// TestCreateRoomScansReturnedRecord verifies room creation scans records.
func TestCreateRoomScansReturnedRecord(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: roomValuesForTest()}}
	room, err := New(executor).CreateRoom(context.Background(), CreateRoomParams{OwnerPlayerID: 7, OwnerName: "demo", Name: "Test Room", ModelName: "model_a", MaxUsers: 25})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	if room.ID != 9 || !strings.Contains(executor.query, "insert into rooms") {
		t.Fatalf("unexpected room=%#v query=%q", room, executor.query)
	}
}

// TestFindRoomByIDReportsMissing verifies missing room lookup.
func TestFindRoomByIDReportsMissing(t *testing.T) {
	_, found, err := New(&fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}).FindRoomByID(context.Background(), 9)
	if err != nil {
		t.Fatalf("find room: %v", err)
	}
	if found {
		t.Fatal("expected missing room")
	}
}

// TestListRoomsByOwnerScansRows verifies room list scanning.
func TestListRoomsByOwnerScansRows(t *testing.T) {
	rooms, err := New(&fakeExecutor{rows: &fakeRows{values: [][]any{roomValuesForTest()}}}).ListRoomsByOwner(context.Background(), 7)
	if err != nil {
		t.Fatalf("list rooms: %v", err)
	}
	if len(rooms) != 1 || rooms[0].Name != "Test Room" {
		t.Fatalf("unexpected rooms %#v", rooms)
	}
}

// TestListPopularRoomsUsesQuery verifies popular room query path.
func TestListPopularRoomsUsesQuery(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{roomValuesForTest()}}}
	_, err := New(executor).ListPopularRooms(context.Background(), 10)
	if err != nil {
		t.Fatalf("list popular rooms: %v", err)
	}
	if !strings.Contains(executor.query, "order by score desc") {
		t.Fatalf("unexpected query %q", executor.query)
	}
}

// TestListHighestScoreRoomsUsesQuery verifies highest score query path.
func TestListHighestScoreRoomsUsesQuery(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{roomValuesForTest()}}}
	_, err := New(executor).ListHighestScoreRooms(context.Background(), 10)
	if err != nil {
		t.Fatalf("list highest score rooms: %v", err)
	}
	if !strings.Contains(executor.query, "id asc") {
		t.Fatalf("unexpected query %q", executor.query)
	}
}

// TestSoftDeleteRoomReportsRowsAffected verifies soft delete behavior.
func TestSoftDeleteRoomReportsRowsAffected(t *testing.T) {
	deleted, err := New(&fakeExecutor{tag: pgconn.NewCommandTag("UPDATE 1")}).SoftDeleteRoom(context.Background(), 9)
	if err != nil {
		t.Fatalf("soft delete room: %v", err)
	}
	if !deleted {
		t.Fatal("expected deleted room")
	}
}

// TestReplaceRoomTagsExecutesDeleteAndInsert verifies tag replacement.
func TestReplaceRoomTagsExecutesDeleteAndInsert(t *testing.T) {
	executor := &fakeExecutor{}
	err := New(executor).ReplaceRoomTags(context.Background(), 9, []string{"fun", "social"})
	if err != nil {
		t.Fatalf("replace tags: %v", err)
	}
	if executor.execs != 3 {
		t.Fatalf("expected three execs, got %d", executor.execs)
	}
}

// TestListRoomTagsScansRows verifies room tag scanning.
func TestListRoomTagsScansRows(t *testing.T) {
	tags, err := New(&fakeExecutor{rows: &fakeRows{values: [][]any{{int64(9), "fun"}}}}).ListRoomTags(context.Background(), 9)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 1 || tags[0].Value != "fun" {
		t.Fatalf("unexpected tags %#v", tags)
	}
}

// TestListCategoriesScansRows verifies category scanning.
func TestListCategoriesScansRows(t *testing.T) {
	categories, err := New(&fakeExecutor{rows: &fakeRows{values: [][]any{categoryValuesForTest()}}}).ListCategories(context.Background())
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	if len(categories) != 1 || categories[0].Caption != "Social" {
		t.Fatalf("unexpected categories %#v", categories)
	}
}

// TestListCategoriesWrapsRowsError verifies category row errors.
func TestListCategoriesWrapsRowsError(t *testing.T) {
	expected := errors.New("rows failed")
	_, err := New(&fakeExecutor{rows: &fakeRows{err: expected}}).ListCategories(context.Background())
	if !errors.Is(err, expected) {
		t.Fatalf("expected rows error, got %v", err)
	}
}

// roomValuesForTest returns scannable room values.
func roomValuesForTest() []any {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	return []any{int64(9), int64(7), "demo", "Test Room", "hello", "model_a", int16(roommodel.DoorModeOpen), 25, 3, pgtype.Int8{}, int16(roommodel.TradeModeDisabled), false, true, false, false, 0, 0, int16(0), int16(1), int16(1), int16(50), int16(2), false, false, now, now, pgtype.Timestamptz{}, int64(1)}
}

// categoryValuesForTest returns scannable category values.
func categoryValuesForTest() []any {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	return []any{int64(1), "Social", "social", true, false, "", "", false, 1, now, now, pgtype.Timestamptz{}, int64(1)}
}
