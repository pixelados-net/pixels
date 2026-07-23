package entrytile

import (
	"context"
	"errors"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
)

// roomManagerForTest provides room lookup fixtures.
type roomManagerForTest struct {
	// room stores the room lookup result.
	room roommodel.Room
	// found reports whether the room exists.
	found bool
	// err stores the room lookup error.
	err error
}

// Create returns no room for entry tile tests.
func (manager roomManagerForTest) Create(context.Context, roomservice.CreateParams) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// FindByID finds a fixture room by id.
func (manager roomManagerForTest) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return manager.room, manager.found, manager.err
}

// ListByOwner returns no rooms for entry tile tests.
func (manager roomManagerForTest) ListByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return nil, nil
}

// ListPopular returns no rooms for entry tile tests.
func (manager roomManagerForTest) ListPopular(context.Context, int) ([]roommodel.Room, error) {
	return nil, nil
}

// ListHighestScore returns no rooms for entry tile tests.
func (manager roomManagerForTest) ListHighestScore(context.Context, int) ([]roommodel.Room, error) {
	return nil, nil
}

// Search returns no rooms for entry tile tests.
func (manager roomManagerForTest) Search(context.Context, string, int) ([]roommodel.Room, error) {
	return nil, nil
}

// ListTags returns no tags for entry tile tests.
func (manager roomManagerForTest) ListTags(context.Context, int64) ([]roommodel.Tag, error) {
	return nil, nil
}

// SoftDelete deletes nothing for entry tile tests.
func (manager roomManagerForTest) SoftDelete(context.Context, int64) error {
	return nil
}

// ListCategories returns no categories for entry tile tests.
func (manager roomManagerForTest) ListCategories(context.Context) ([]roommodel.Category, error) {
	return nil, nil
}

// layoutManagerForTest provides layout lookup fixtures.
type layoutManagerForTest struct {
	// roomLayout stores the layout lookup result.
	roomLayout layout.Layout
	// found reports whether the layout exists.
	found bool
	// err stores the layout lookup error.
	err error
}

// Create returns no layout for entry tile tests.
func (manager layoutManagerForTest) Create(context.Context, layout.SaveParams) (layout.Layout, error) {
	return layout.Layout{}, nil
}

// Update returns no layout for entry tile tests.
func (manager layoutManagerForTest) Update(context.Context, int64, layout.SaveParams) (layout.Layout, error) {
	return layout.Layout{}, nil
}

// FindByID returns no layout for entry tile tests.
func (manager layoutManagerForTest) FindByID(context.Context, int64) (layout.Layout, bool, error) {
	return layout.Layout{}, false, nil
}

// FindByName finds a fixture layout by name.
func (manager layoutManagerForTest) FindByName(context.Context, string) (layout.Layout, bool, error) {
	return manager.roomLayout, manager.found, manager.err
}

// errLookupFailed stores a reusable lookup failure.
var errLookupFailed = errors.New("lookup failed")

// List returns no layouts for entry tile tests.
func (manager layoutManagerForTest) List(context.Context) ([]layout.Layout, error) {
	return nil, nil
}

// Catalog returns no catalog for entry tile tests.
func (manager layoutManagerForTest) Catalog(context.Context) (*layout.Catalog, error) {
	return nil, nil
}
