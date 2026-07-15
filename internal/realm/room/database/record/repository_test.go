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
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestCreateRoomScansReturnedRecord verifies room creation scans records.
func TestCreateRoomScansReturnedRecord(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: roomValuesForTest()}}
	room, err := New(executor).CreateRoom(context.Background(), roomservice.CreateRecordParams{OwnerPlayerID: 7, OwnerName: "demo", Name: "Test Room", ModelName: "model_a", MaxUsers: 25})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	if room.ID != 9 || room.PasswordHash == nil || *room.PasswordHash != "bcrypt-hash" || !strings.Contains(executor.query, "insert into rooms") {
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
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{roomValuesForTest()}}}
	rooms, err := New(executor).ListRoomsByOwner(context.Background(), 7)
	if err != nil {
		t.Fatalf("list rooms: %v", err)
	}
	if len(rooms) != 1 || rooms[0].Name != "Test Room" {
		t.Fatalf("unexpected rooms %#v", rooms)
	}
	if !strings.Contains(executor.query, "not is_bundle_template") {
		t.Fatalf("template rooms must not count as owned rooms: %q", executor.query)
	}
}

// TestListPopularRoomsUsesQuery verifies popular room query path.
func TestListPopularRoomsUsesQuery(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{roomValuesForTest()}}}
	_, err := New(executor).ListPopularRooms(context.Background(), 10)
	if err != nil {
		t.Fatalf("list popular rooms: %v", err)
	}
	if !strings.Contains(executor.query, "door_mode <> 3") || !strings.Contains(executor.query, "not is_bundle_template") || !strings.Contains(executor.query, "order by score desc") {
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
	if !strings.Contains(executor.query, "door_mode <> 3") || !strings.Contains(executor.query, "not is_bundle_template") || !strings.Contains(executor.query, "id asc") {
		t.Fatalf("unexpected query %q", executor.query)
	}
}

// TestSearchRoomsExcludesInvisibleRooms verifies public text search visibility.
func TestSearchRoomsExcludesInvisibleRooms(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{roomValuesForTest()}}}
	if _, err := New(executor).SearchRooms(context.Background(), "pixels", 10); err != nil {
		t.Fatalf("search rooms: %v", err)
	}
	if !strings.Contains(executor.query, "door_mode <> 3") || !strings.Contains(executor.query, "not is_bundle_template") {
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

// TestUpdateRoomWritesSettingsAndReplacesTags verifies the atomic mutation work sequence.
func TestUpdateRoomWritesSettingsAndReplacesTags(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: roomValuesForTest()}}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, Name: "Updated", MaxUsers: 25}
	updated, found, err := New(executor).UpdateRoom(context.Background(), roomservice.UpdateRecordParams{Room: room, ExpectedVersion: 1}, []string{"social", "build"})
	if err != nil || !found {
		t.Fatalf("updated=%#v found=%v err=%v", updated, found, err)
	}
	if executor.execs != 3 || !strings.Contains(executor.query, "version=$2") {
		t.Fatalf("execs=%d query=%q", executor.execs, executor.query)
	}
}

// TestUpdateRoomReportsVersionConflict verifies missing optimistic rows skip tag replacement.
func TestUpdateRoomReportsVersionConflict(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}
	_, found, err := New(executor).UpdateRoom(context.Background(), roomservice.UpdateRecordParams{Room: roommodel.Room{}, ExpectedVersion: 99}, []string{"social"})
	if err != nil || found || executor.execs != 0 {
		t.Fatalf("found=%v execs=%d err=%v", found, executor.execs, err)
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

// TestLockRoomOwnerTypesPlayerIDAsBigint verifies pgx can encode the lock argument.
func TestLockRoomOwnerTypesPlayerIDAsBigint(t *testing.T) {
	executor := &fakeExecutor{}
	if err := New(executor).LockRoomOwner(context.Background(), 7); err != nil {
		t.Fatalf("lock room owner: %v", err)
	}
	if !strings.Contains(executor.execQuery, "$1::bigint") {
		t.Fatalf("owner lock must type its int64 parameter: %q", executor.execQuery)
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
	return []any{int64(9), int64(7), "demo", "Test Room", "hello", "model_a", int16(roommodel.DoorModeOpen), pgtype.Text{String: "bcrypt-hash", Valid: true}, 25, 3, pgtype.Int8{}, int16(roommodel.TradeModeDisabled), 4, false, true, false, false, 0, 0, int16(0), int16(1), int16(1), int16(50), int16(2), int16(0), int16(1), int16(2), false, false, false, "0.0", "0.0", "0.0", now, now, pgtype.Timestamptz{}, int64(1)}
}

// categoryValuesForTest returns scannable category values.
func categoryValuesForTest() []any {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	return []any{int64(1), "Social", "social", true, false, "", "", false, 1, now, now, pgtype.Timestamptz{}, int64(1)}
}
