package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// TestFindItemByIDScansPlacedRecord verifies item lookup scanning for a placed item.
func TestFindItemByIDScansPlacedRecord(t *testing.T) {
	item, found, err := New(&fakeExecutor{row: fakeRow{values: itemValuesForTest(placedRoomID, placedX, placedY, placedZ)}}).FindItemByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("find item: %v", err)
	}
	if !found || !item.InRoom() || item.InInventory() {
		t.Fatalf("unexpected placed item=%#v found=%v", item, found)
	}
	if *item.RoomID != 1 || *item.X != 4 || *item.Y != 4 || *item.Z != 0 {
		t.Fatalf("unexpected placement %#v", item)
	}
}

// TestFindItemByIDScansInventoryRecord verifies item lookup scanning for an inventory item.
func TestFindItemByIDScansInventoryRecord(t *testing.T) {
	item, found, err := New(&fakeExecutor{row: fakeRow{values: itemValuesForTest(pgtype.Int8{}, pgtype.Int2{}, pgtype.Int2{}, pgtype.Float8{})}}).FindItemByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("find item: %v", err)
	}
	if !found || item.InRoom() || !item.InInventory() {
		t.Fatalf("unexpected inventory item=%#v found=%v", item, found)
	}
	if item.RoomID != nil || item.X != nil || item.Y != nil || item.Z != nil {
		t.Fatalf("expected null placement fields %#v", item)
	}
}

// TestFindItemByIDReportsMissing verifies missing item lookup.
func TestFindItemByIDReportsMissing(t *testing.T) {
	_, found, err := New(&fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}).FindItemByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("find item: %v", err)
	}
	if found {
		t.Fatal("expected missing item")
	}
}

// TestListInventoryItemsUsesOwnerQuery verifies inventory listing query shape.
func TestListInventoryItemsUsesOwnerQuery(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{itemValuesForTest(pgtype.Int8{}, pgtype.Int2{}, pgtype.Int2{}, pgtype.Float8{})}}}
	items, err := New(executor).ListInventoryItems(context.Background(), 7)
	if err != nil {
		t.Fatalf("list inventory items: %v", err)
	}
	if len(items) != 1 || !items[0].InInventory() {
		t.Fatalf("unexpected inventory items %#v", items)
	}
	if !strings.Contains(executor.query, "room_id is null") {
		t.Fatalf("unexpected query %q", executor.query)
	}
}

// TestListRoomItemsUsesRoomQuery verifies room listing query shape.
func TestListRoomItemsUsesRoomQuery(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{itemValuesForTest(placedRoomID, placedX, placedY, placedZ)}}}
	items, err := New(executor).ListRoomItems(context.Background(), 1)
	if err != nil {
		t.Fatalf("list room items: %v", err)
	}
	if len(items) != 1 || !items[0].InRoom() {
		t.Fatalf("unexpected room items %#v", items)
	}
	if !strings.Contains(executor.query, "where room_id = $1") {
		t.Fatalf("unexpected query %q", executor.query)
	}
}

// TestPlaceItemUsesInventoryGuardedQuery verifies placement query shape and arguments.
func TestPlaceItemUsesInventoryGuardedQuery(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: itemValuesForTest(placedRoomID, placedX, placedY, placedZ)}}
	item, updated, err := New(executor).PlaceItem(context.Background(), PlaceItemParams{
		ID:            1,
		OwnerPlayerID: 7,
		RoomID:        1,
		Placement:     furnituremodel.Placement{X: 4, Y: 4, Z: 0, Rotation: furnituremodel.RotationNorth},
	})
	if err != nil {
		t.Fatalf("place item: %v", err)
	}
	if !updated || !item.InRoom() {
		t.Fatalf("unexpected place result item=%#v updated=%v", item, updated)
	}
	if !strings.Contains(executor.query, "room_id is null") || !strings.Contains(executor.query, "$4::smallint") || !strings.Contains(executor.query, "$6::numeric(6,2)") || len(executor.arguments) != 9 {
		t.Fatalf("unexpected query %q arguments=%#v", executor.query, executor.arguments)
	}
}

// TestPlaceWallItemUsesTypedNullableCoordinates verifies wall placement avoids PostgreSQL CASE type ambiguity.
func TestPlaceWallItemUsesTypedNullableCoordinates(t *testing.T) {
	values := itemValuesForTest(placedRoomID, pgtype.Int2{}, pgtype.Int2{}, pgtype.Float8{})
	values[8] = pgtype.Text{String: ":w=3,4 l=13,22 l", Valid: true}
	executor := &fakeExecutor{row: fakeRow{values: values}}
	item, updated, err := New(executor).PlaceItem(context.Background(), PlaceItemParams{
		ID: 1, OwnerPlayerID: 7, RoomID: 1, WallPosition: ":w=3,4 l=13,22 l",
	})
	if err != nil {
		t.Fatalf("place wall item: %v", err)
	}
	if !updated || !item.InRoom() || item.WallPosition == nil {
		t.Fatalf("unexpected wall placement item=%#v updated=%v", item, updated)
	}
	if !strings.Contains(executor.query, "$8::text") || !strings.Contains(executor.query, "null::smallint") {
		t.Fatalf("expected explicitly typed wall query %q", executor.query)
	}
}

// TestPlaceItemReportsNoRowsAsNotUpdated verifies placement conflict handling.
func TestPlaceItemReportsNoRowsAsNotUpdated(t *testing.T) {
	_, updated, err := New(&fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}).PlaceItem(context.Background(), PlaceItemParams{ID: 1, OwnerPlayerID: 7, RoomID: 1})
	if err != nil {
		t.Fatalf("place item: %v", err)
	}
	if updated {
		t.Fatal("expected no rows updated")
	}
}

// TestMoveItemUsesPlacedGuardedQuery verifies move query shape.
func TestMoveItemUsesPlacedGuardedQuery(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: itemValuesForTest(placedRoomID, placedX, placedY, placedZ)}}
	item, updated, err := New(executor).MoveItem(context.Background(), MoveItemParams{
		ID:        1,
		RoomID:    1,
		Placement: furnituremodel.Placement{X: 5, Y: 5, Z: 0, Rotation: furnituremodel.RotationEast},
	})
	if err != nil {
		t.Fatalf("move item: %v", err)
	}
	if !updated || !item.InRoom() {
		t.Fatalf("unexpected move result item=%#v updated=%v", item, updated)
	}
	if !strings.Contains(executor.query, "room_id = $2") || !strings.Contains(executor.query, "$3::smallint") || !strings.Contains(executor.query, "$5::numeric(6,2)") || len(executor.arguments) != 7 || executor.arguments[1] != int64(1) {
		t.Fatalf("unexpected query %q arguments=%#v", executor.query, executor.arguments)
	}
}

// TestPickupItemUsesPlacedGuardedQuery verifies pickup query shape.
func TestPickupItemUsesPlacedGuardedQuery(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: itemValuesForTest(pgtype.Int8{}, pgtype.Int2{}, pgtype.Int2{}, pgtype.Float8{})}}
	item, updated, err := New(executor).PickupItem(context.Background(), PickupItemParams{ID: 1, OwnerPlayerID: 7})
	if err != nil {
		t.Fatalf("pickup item: %v", err)
	}
	if !updated || !item.InInventory() {
		t.Fatalf("unexpected pickup result item=%#v updated=%v", item, updated)
	}
	if !strings.Contains(executor.query, "room_id = null") || len(executor.arguments) != 2 {
		t.Fatalf("unexpected query %q arguments=%#v", executor.query, executor.arguments)
	}
}

// TestPickupItemReportsNoRowsAsNotUpdated verifies pickup conflict handling.
func TestPickupItemReportsNoRowsAsNotUpdated(t *testing.T) {
	_, updated, err := New(&fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}).PickupItem(context.Background(), PickupItemParams{ID: 1, OwnerPlayerID: 7})
	if err != nil {
		t.Fatalf("pickup item: %v", err)
	}
	if updated {
		t.Fatal("expected no rows updated")
	}
}

// TestUpdateItemStateUsesCompareAndSwap verifies state query guards and arguments.
func TestUpdateItemStateUsesCompareAndSwap(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: itemValuesForTest(placedRoomID, placedX, placedY, placedZ)}}
	_, updated, err := New(executor).UpdateItemState(context.Background(), UpdateItemStateParams{
		ID: 1, RoomID: 9, Expected: "0", Next: "1",
	})
	if err != nil {
		t.Fatalf("update item state: %v", err)
	}
	if !updated || !strings.Contains(executor.query, "extra_data = $3") || len(executor.arguments) != 4 {
		t.Fatalf("unexpected state query %q arguments=%#v", executor.query, executor.arguments)
	}
}

// TestUpdateItemStateReportsConflict verifies a failed compare-and-swap.
func TestUpdateItemStateReportsConflict(t *testing.T) {
	_, updated, err := New(&fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}).UpdateItemState(context.Background(), UpdateItemStateParams{ID: 1, RoomID: 9})
	if err != nil || updated {
		t.Fatalf("expected unmatched state update updated=%v err=%v", updated, err)
	}
}

var (
	// placedRoomID stores a placed item room id fixture.
	placedRoomID = pgtype.Int8{Int64: 1, Valid: true}

	// placedX stores a placed item x fixture.
	placedX = pgtype.Int2{Int16: 4, Valid: true}

	// placedY stores a placed item y fixture.
	placedY = pgtype.Int2{Int16: 4, Valid: true}

	// placedZ stores a placed item z fixture.
	placedZ = pgtype.Float8{Float64: 0, Valid: true}
)

// definitionValuesForTest returns scannable furniture definition values.
func definitionValuesForTest() []any {
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	return []any{
		int64(2), 39, "chair_plasto", "chair_plasto", "floor", 1, 1, 1.0,
		false, false, true, false, true, true, true, int32(0), []int32{101, 102}, pgtype.Int4{}, pgtype.Int4{},
		"default", 2, "", "",
		[]byte(`{"slots":[{"dx":0,"dy":0,"status":"sit","body_rotation":4}]}`),
		now, now, pgtype.Timestamptz{}, int64(1),
	}
}

// itemValuesForTest returns scannable furniture item values for a given placement state.
func itemValuesForTest(roomID pgtype.Int8, x pgtype.Int2, y pgtype.Int2, z pgtype.Float8) []any {
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	return []any{
		int64(1), int64(2), int64(7), roomID, x, y, z, int16(0),
		pgtype.Text{}, "0", pgtype.Int4{}, false, false, pgtype.Int4{}, pgtype.Int4{}, pgtype.Int4{}, pgtype.Int8{}, pgtype.Text{}, []byte(`{}`),
		now, now, pgtype.Timestamptz{}, int64(1),
	}
}
