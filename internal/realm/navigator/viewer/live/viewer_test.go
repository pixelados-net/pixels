package live

import "testing"

// TestViewerStoresLastSearch verifies viewer search state.
func TestViewerStoresLastSearch(t *testing.T) {
	viewer := NewViewer()
	viewer.SetLastSearch(LastSearch{Code: "hotel_view", Query: "demo"})

	search := viewer.LastSearch()
	if search.Code != "hotel_view" || search.Query != "demo" {
		t.Fatalf("unexpected search %#v", search)
	}
}

// TestViewerStoresVisibleRooms verifies visible room snapshots.
func TestViewerStoresVisibleRooms(t *testing.T) {
	viewer := NewViewer()
	viewer.SetSearch(LastSearch{Code: "hotel_view"}, []int64{4, 7, 7, 0})

	if !viewer.HasVisibleRoom(4) || !viewer.HasVisibleRoom(7) {
		t.Fatalf("expected visible rooms %#v", viewer.VisibleRoomIDs())
	}
	if viewer.HasVisibleRoom(9) {
		t.Fatal("unexpected visible room")
	}
}

// TestViewerMatchesAnyVisibleRoom verifies room set intersections.
func TestViewerMatchesAnyVisibleRoom(t *testing.T) {
	viewer := NewViewer()
	viewer.SetSearch(LastSearch{Code: "hotel_view"}, []int64{4})

	if !viewer.HasAnyVisibleRoom(map[int64]struct{}{3: {}, 4: {}}) {
		t.Fatal("expected matching visible room")
	}
	if viewer.HasAnyVisibleRoom(map[int64]struct{}{9: {}}) {
		t.Fatal("unexpected matching visible room")
	}
}

// TestViewerCategoryCounts verifies category count preference state.
func TestViewerCategoryCounts(t *testing.T) {
	viewer := NewViewer()
	if !viewer.ReceivesCategoryCounts() {
		t.Fatal("expected category counts by default")
	}

	viewer.SetCategoryCounts(false)
	if viewer.ReceivesCategoryCounts() {
		t.Fatal("expected category counts disabled")
	}
}
