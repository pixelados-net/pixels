package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// TestAddFavoriteExecutesInsert verifies favorite insertion.
func TestAddFavoriteExecutesInsert(t *testing.T) {
	executor := &fakeExecutor{}
	err := New(executor).AddFavorite(context.Background(), 7, 9)
	if err != nil {
		t.Fatalf("add favorite: %v", err)
	}
	if executor.execs != 1 {
		t.Fatalf("expected one exec, got %d", executor.execs)
	}
}

// TestRemoveFavoriteExecutesDelete verifies favorite removal.
func TestRemoveFavoriteExecutesDelete(t *testing.T) {
	executor := &fakeExecutor{}
	err := New(executor).RemoveFavorite(context.Background(), 7, 9)
	if err != nil {
		t.Fatalf("remove favorite: %v", err)
	}
	if executor.execs != 1 {
		t.Fatalf("expected one exec, got %d", executor.execs)
	}
}

// TestListFavoriteRoomIDsScansRows verifies favorite listing.
func TestListFavoriteRoomIDsScansRows(t *testing.T) {
	ids, err := New(&fakeExecutor{rows: &fakeRows{values: [][]any{{int64(9)}}}}).ListFavoriteRoomIDs(context.Background(), 7)
	if err != nil {
		t.Fatalf("list favorites: %v", err)
	}
	if len(ids) != 1 || ids[0] != 9 {
		t.Fatalf("unexpected ids %#v", ids)
	}
}

// TestSaveSearchScansRecord verifies saved search scanning.
func TestSaveSearchScansRecord(t *testing.T) {
	search, err := New(&fakeExecutor{row: fakeRow{values: savedSearchValuesForTest()}}).SaveSearch(context.Background(), SaveSearchParams{PlayerID: 7, Code: "hotel_view"})
	if err != nil {
		t.Fatalf("save search: %v", err)
	}
	if search.ID != 3 || search.Code != "hotel_view" {
		t.Fatalf("unexpected search %#v", search)
	}
}

// TestDeleteSearchReportsRowsAffected verifies delete behavior.
func TestDeleteSearchReportsRowsAffected(t *testing.T) {
	deleted, err := New(&fakeExecutor{tag: pgconn.NewCommandTag("DELETE 1")}).DeleteSearch(context.Background(), 7, 3)
	if err != nil {
		t.Fatalf("delete search: %v", err)
	}
	if !deleted {
		t.Fatal("expected deleted search")
	}
}

// TestListSavedSearchesScansRows verifies saved search listing.
func TestListSavedSearchesScansRows(t *testing.T) {
	searches, err := New(&fakeExecutor{rows: &fakeRows{values: [][]any{savedSearchValuesForTest()}}}).ListSavedSearches(context.Background(), 7)
	if err != nil {
		t.Fatalf("list searches: %v", err)
	}
	if len(searches) != 1 || searches[0].ID != 3 {
		t.Fatalf("unexpected searches %#v", searches)
	}
}

// TestFindPreferenceReportsMissing verifies missing preference lookup.
func TestFindPreferenceReportsMissing(t *testing.T) {
	_, found, err := New(&fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}).FindPreference(context.Background(), 7)
	if err != nil {
		t.Fatalf("find preference: %v", err)
	}
	if found {
		t.Fatal("expected missing preference")
	}
}

// TestFindPreferenceScansRecord verifies preference lookup.
func TestFindPreferenceScansRecord(t *testing.T) {
	preference, found, err := New(&fakeExecutor{row: fakeRow{values: preferenceValuesForTest()}}).FindPreference(context.Background(), 7)
	if err != nil {
		t.Fatalf("find preference: %v", err)
	}
	if !found || preference.PlayerID != 7 {
		t.Fatalf("unexpected preference %#v found=%v", preference, found)
	}
}

// TestUpsertPreferenceScansRecord verifies preference upsert scanning.
func TestUpsertPreferenceScansRecord(t *testing.T) {
	preference, err := New(&fakeExecutor{row: fakeRow{values: preferenceValuesForTest()}}).UpsertPreference(context.Background(), PreferenceParams{})
	if err != nil {
		t.Fatalf("upsert preference: %v", err)
	}
	if preference.PlayerID != 7 || preference.WindowWidth != 425 {
		t.Fatalf("unexpected preference %#v", preference)
	}
}

// TestUpsertCategoryPreferenceScansRecord verifies category preference upsert.
func TestUpsertCategoryPreferenceScansRecord(t *testing.T) {
	preference, err := New(&fakeExecutor{row: fakeRow{values: categoryPreferenceValuesForTest()}}).UpsertCategoryPreference(context.Background(), CategoryPreferenceParams{})
	if err != nil {
		t.Fatalf("upsert category preference: %v", err)
	}
	if preference.Code != "popular" {
		t.Fatalf("unexpected preference %#v", preference)
	}
}

// TestListCategoryPreferencesScansRows verifies category preference listing.
func TestListCategoryPreferencesScansRows(t *testing.T) {
	preferences, err := New(&fakeExecutor{rows: &fakeRows{values: [][]any{categoryPreferenceValuesForTest()}}}).ListCategoryPreferences(context.Background(), 7)
	if err != nil {
		t.Fatalf("list category preferences: %v", err)
	}
	if len(preferences) != 1 || preferences[0].Code != "popular" {
		t.Fatalf("unexpected preferences %#v", preferences)
	}
}

// TestListLiftedRoomsScansRows verifies lifted room scanning.
func TestListLiftedRoomsScansRows(t *testing.T) {
	rooms, err := New(&fakeExecutor{rows: &fakeRows{values: [][]any{liftedValuesForTest()}}}).ListLiftedRooms(context.Background())
	if err != nil {
		t.Fatalf("list lifted: %v", err)
	}
	if len(rooms) != 1 || rooms[0].RoomID != 9 {
		t.Fatalf("unexpected lifted rooms %#v", rooms)
	}
}

// TestListSavedSearchesWrapsRowsError verifies row errors.
func TestListSavedSearchesWrapsRowsError(t *testing.T) {
	expected := errors.New("rows failed")
	_, err := New(&fakeExecutor{rows: &fakeRows{err: expected}}).ListSavedSearches(context.Background(), 7)
	if !errors.Is(err, expected) {
		t.Fatalf("expected rows error, got %v", err)
	}
}

// savedSearchValuesForTest returns scannable saved search values.
func savedSearchValuesForTest() []any {
	return []any{int64(3), int64(7), "hotel_view", "demo", "", time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)}
}

// liftedValuesForTest returns scannable lifted room values.
func liftedValuesForTest() []any {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	return []any{int64(1), int64(9), 0, "image.png", "Lobby", 1, pgtype.Timestamptz{}, pgtype.Timestamptz{}, now, now, pgtype.Timestamptz{}, int64(1)}
}

// preferenceValuesForTest returns scannable preference values.
func preferenceValuesForTest() []any {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	return []any{int64(7), 68, 42, 425, 592, false, int16(0), now, now}
}

// categoryPreferenceValuesForTest returns scannable category preference values.
func categoryPreferenceValuesForTest() []any {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	return []any{int64(7), "popular", true, int16(1), now, now}
}
