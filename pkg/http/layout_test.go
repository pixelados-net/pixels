package http

import (
	"context"

	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
)

// testLayouts returns HTTP test room layouts.
func testLayouts() roomlayout.Manager { return testLayoutManager{} }

// testLayoutManager provides room layouts for HTTP tests.
type testLayoutManager struct{}

// Create creates no layout.
func (testLayoutManager) Create(context.Context, roomlayout.SaveParams) (roomlayout.Layout, error) {
	return roomlayout.Layout{}, nil
}

// Update updates no layout.
func (testLayoutManager) Update(context.Context, int64, roomlayout.SaveParams) (roomlayout.Layout, error) {
	return roomlayout.Layout{}, nil
}

// FindByID finds no layout.
func (testLayoutManager) FindByID(context.Context, int64) (roomlayout.Layout, bool, error) {
	return roomlayout.Layout{}, false, nil
}

// FindByName finds no layout.
func (testLayoutManager) FindByName(context.Context, string) (roomlayout.Layout, bool, error) {
	return roomlayout.Layout{}, false, nil
}

// List lists one enabled layout.
func (testLayoutManager) List(context.Context) ([]roomlayout.Layout, error) {
	return []roomlayout.Layout{{Name: "model_a", TileSize: 104, Enabled: true}}, nil
}

// Catalog returns no layout catalog.
func (testLayoutManager) Catalog(context.Context) (*roomlayout.Catalog, error) { return nil, nil }
