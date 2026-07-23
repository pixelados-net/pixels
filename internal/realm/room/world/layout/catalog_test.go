package layout

import (
	"errors"
	"testing"
)

// TestCatalogFindNormalizesNames verifies client and protocol model names.
func TestCatalogFindNormalizesNames(t *testing.T) {
	catalog := catalogForTest(t)

	roomLayout, found := catalog.Find("a")
	if !found {
		t.Fatal("expected default model a")
	}
	if roomLayout.Name != "model_a" || roomLayout.TileSize != 12 {
		t.Fatalf("unexpected layout %#v", roomLayout)
	}
}

// TestCatalogRejectsInvalidLayouts verifies registration validation.
func TestCatalogRejectsInvalidLayouts(t *testing.T) {
	_, err := NewCatalog([]Layout{{Name: "model_empty"}})
	if !errors.Is(err, ErrInvalidLayout) {
		t.Fatalf("expected invalid layout error, got %v", err)
	}
}

// TestCatalogMustFindReportsMissingLayout verifies missing layout errors.
func TestCatalogMustFindReportsMissingLayout(t *testing.T) {
	catalog := catalogForTest(t)

	_, err := catalog.MustFind("missing")
	if !errors.Is(err, ErrLayoutNotFound) {
		t.Fatalf("expected layout not found error, got %v", err)
	}
}

// TestCatalogListReturnsRegisteredLayouts verifies list behavior.
func TestCatalogListReturnsRegisteredLayouts(t *testing.T) {
	catalog := catalogForTest(t)

	layouts := catalog.List()
	if len(layouts) != 1 {
		t.Fatalf("expected 1 layout, got %d", len(layouts))
	}
}

// catalogForTest creates a catalog for tests.
func catalogForTest(t *testing.T) *Catalog {
	t.Helper()

	catalog, err := NewCatalog([]Layout{validLayoutForTest()})
	if err != nil {
		t.Fatalf("create catalog: %v", err)
	}

	return catalog
}
