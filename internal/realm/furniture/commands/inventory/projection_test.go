package inventory

import (
	"testing"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/furniture/list"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestInventoryCategoryMapsRoomEffects verifies special room consumable categories.
func TestInventoryCategoryMapsRoomEffects(t *testing.T) {
	tests := map[string]outlist.Category{
		"chair": outlist.CategoryDefault, "wallpaper": outlist.CategoryWallpaper,
		"floor": outlist.CategoryFloor, "landscape": outlist.CategoryLandscape,
	}
	for name, expected := range tests {
		if actual := inventoryCategory(name); actual != expected {
			t.Fatalf("inventoryCategory(%q) = %d, want %d", name, actual, expected)
		}
	}
}

// TestFragmentRecordsProjectsGiftWrapper verifies unopened gifts use wrapper visuals.
func TestFragmentRecordsProjectsGiftWrapper(t *testing.T) {
	sprite, box, ribbon := int32(3379), int32(2), int32(7)
	items := []furnituremodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 40}}, DefinitionID: 2,
		GiftWrapped: true, GiftWrapSpriteID: &sprite, GiftWrapBoxID: &box, GiftWrapRibbonID: &ribbon}}
	definitions := map[int64]furnituremodel.Definition{2: {SpriteID: 39, Kind: furnituremodel.KindFloor}}
	records := fragmentRecords(items, definitions, nil, 0)
	if len(records) != 1 || records[0].SpriteID != 3379 || records[0].GiftBoxID != 2 || records[0].GiftRibbonID != 7 || records[0].AllowInventoryStack {
		t.Fatalf("unexpected gift inventory records %#v", records)
	}
}

// TestFragmentRecordsProjectsGroupObjectData verifies warmed group identity reaches Nitro's inventory model.
func TestFragmentRecordsProjectsGroupObjectData(t *testing.T) {
	items := []furnituremodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 41}}, DefinitionID: 2, ExtraData: "1"}}
	definitions := map[int64]furnituremodel.Definition{2: {SpriteID: 4254, Kind: furnituremodel.KindFloor}}
	groups := groupPolicyForTest{items: map[int64]furnituremodel.GroupData{41: {
		GroupID: 2, BadgeCode: "b14014s05050", ColorAHex: "#ff0000", ColorBHex: "#00ff00",
	}}}

	records := fragmentRecords(items, definitions, groups, 0)
	if len(records) != 1 || records[0].Data == nil {
		t.Fatalf("unexpected group inventory records %#v", records)
	}
	expected := []string{"1", "2", "b14014s05050", "#ff0000", "#00ff00"}
	if len(records[0].Data.Strings) != len(expected) {
		t.Fatalf("unexpected group object data %#v", records[0].Data.Strings)
	}
	for index := range expected {
		if records[0].Data.Strings[index] != expected[index] {
			t.Fatalf("group object data[%d]=%q want %q", index, records[0].Data.Strings[index], expected[index])
		}
	}
}

// groupPolicyForTest resolves group inventory fixtures.
type groupPolicyForTest struct {
	// items stores group identity by furniture item id.
	items map[int64]furnituremodel.GroupData
}

// Furniture returns one group inventory fixture.
func (policy groupPolicyForTest) Furniture(itemID int64) (furnituremodel.GroupData, bool) {
	value, found := policy.items[itemID]
	return value, found
}
