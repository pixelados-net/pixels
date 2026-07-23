package live

import "testing"

// TestViewerTracksCatalogState verifies transient catalog state.
func TestViewerTracksCatalogState(t *testing.T) {
	viewer := NewViewer()
	if viewer.Mode() != DefaultMode {
		t.Fatalf("unexpected default mode %q", viewer.Mode())
	}
	viewer.SetMode("")
	viewer.SetPage(8)
	pageID, found := viewer.CurrentPage()
	if viewer.Mode() != DefaultMode || !found || pageID != 8 {
		t.Fatalf("unexpected viewer mode=%q page=%d found=%t", viewer.Mode(), pageID, found)
	}
}
