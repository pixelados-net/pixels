package enter

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// furnitureManagerForTest stubs furniture persistence.
type furnitureManagerForTest struct {
	// definitions stores the returned furniture definitions.
	definitions []furnituremodel.Definition

	// items stores the returned room floor items.
	items []furnituremodel.Item

	// err stores the returned error.
	err error
}

// FindDefinitionByID finds a definition by id.
func (manager furnitureManagerForTest) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{}, false, nil
}

// ListDefinitions lists furniture definitions.
func (manager furnitureManagerForTest) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return manager.definitions, manager.err
}

// FindItemByID finds an item by id.
func (manager furnitureManagerForTest) FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error) {
	return furnituremodel.Item{}, false, nil
}

// ListInventory lists inventory items for a player.
func (manager furnitureManagerForTest) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}

// ListRoomItems lists items placed in a room.
func (manager furnitureManagerForTest) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return manager.items, manager.err
}

// Place places an item into a room.
func (manager furnitureManagerForTest) Place(context.Context, furnitureservice.PlaceParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}

// Move repositions a placed item.
func (manager furnitureManagerForTest) Move(context.Context, furnitureservice.MoveParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}

// Pickup returns a placed item to inventory.
func (manager furnitureManagerForTest) Pickup(context.Context, furnitureservice.PickupParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}

// placedFurnitureItemForTest returns a placed furniture item fixture.
func placedFurnitureItemForTest(id int64, definitionID int64, ownerID int64, x int, y int) furnituremodel.Item {
	xValue, yValue, zValue := x, y, 0.0

	return furnituremodel.Item{
		Base:          sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}},
		DefinitionID:  definitionID,
		OwnerPlayerID: ownerID,
		X:             &xValue, Y: &yValue, Z: &zValue,
	}
}

// layoutManagerForTest stubs layout persistence.
type layoutManagerForTest struct {
	// roomLayout stores the returned layout.
	roomLayout layout.Layout

	// found reports whether the layout exists.
	found bool

	// err stores the returned error.
	err error
}

// Create creates a layout record.
func (manager layoutManagerForTest) Create(context.Context, layout.SaveParams) (layout.Layout, error) {
	return layout.Layout{}, nil
}

// Update updates a layout record.
func (manager layoutManagerForTest) Update(context.Context, int64, layout.SaveParams) (layout.Layout, error) {
	return layout.Layout{}, nil
}

// FindByID finds a layout by id.
func (manager layoutManagerForTest) FindByID(context.Context, int64) (layout.Layout, bool, error) {
	return layout.Layout{}, false, nil
}

// FindByName finds a layout by name.
func (manager layoutManagerForTest) FindByName(context.Context, string) (layout.Layout, bool, error) {
	return manager.roomLayout, manager.found, manager.err
}

// List lists layouts.
func (manager layoutManagerForTest) List(context.Context) ([]layout.Layout, error) {
	return nil, nil
}

// Catalog returns layout catalog data.
func (manager layoutManagerForTest) Catalog(context.Context) (*layout.Catalog, error) {
	return nil, nil
}

// playerDirectoryForTest stubs the durable player lookup contract.
type playerDirectoryForTest struct {
	// usernames maps player id to display name for players that exist.
	usernames map[int64]string

	// err stores the returned error.
	err error
}

// FindByID finds a player by id.
func (directory playerDirectoryForTest) FindByID(_ context.Context, id int64) (playerservice.Record, bool, error) {
	if directory.err != nil {
		return playerservice.Record{}, false, directory.err
	}
	username, found := directory.usernames[id]
	if !found {
		return playerservice.Record{}, false, nil
	}

	return playerservice.Record{Player: playermodel.Player{Username: username}}, true, nil
}

// FindByUsername finds a player by username.
func (directory playerDirectoryForTest) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}
