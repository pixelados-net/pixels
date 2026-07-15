package projection

import (
	"testing"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// TestFloorItemsMapsOwnersAndRecords verifies floor item and owner projection.
func TestFloorItemsMapsOwnersAndRecords(t *testing.T) {
	chairX, chairY := 3, 3
	chairZ := 0.0
	bedX, bedY := 5, 6
	bedZ := 0.0

	items := []furnituremodel.Item{
		{DefinitionID: 2, OwnerPlayerID: 1, X: &chairX, Y: &chairY, Z: &chairZ, Rotation: furnituremodel.RotationNorth, ExtraData: "0"},
		{DefinitionID: 5, OwnerPlayerID: 2, X: &bedX, Y: &bedY, Z: &bedZ, Rotation: furnituremodel.RotationSouth, ExtraData: "0"},
	}
	items[0].ID = 10
	items[1].ID = 11
	definitions := map[int64]furnituremodel.Definition{
		2: {SpriteID: 39, AllowSit: true, StackHeight: 1, InteractionModesCount: 2},
		5: {SpriteID: 46, AllowLay: true, StackHeight: 2, InteractionModesCount: 1},
	}
	ownerNames := map[int64]string{1: "demo", 2: "alice"}

	owners, records := FloorItems(items, definitions, ownerNames, nil)

	if len(owners) != 2 || owners[0].ID != 1 || owners[0].Name != "demo" || owners[1].ID != 2 || owners[1].Name != "alice" {
		t.Fatalf("unexpected owners %#v", owners)
	}
	if len(records) != 2 {
		t.Fatalf("unexpected record count %#v", records)
	}

	chair := records[0]
	if chair.ID != 10 || chair.SpriteID != 39 || chair.X != 3 || chair.Y != 3 || chair.Rotation != int(furnituremodel.RotationNorth) {
		t.Fatalf("unexpected chair record %#v", chair)
	}
	if chair.ExtraHeight != "1" || chair.UsagePolicy != 1 || chair.OwnerID != 1 {
		t.Fatalf("unexpected chair sit/usage fields %#v", chair)
	}

	bed := records[1]
	if bed.ExtraHeight != "" {
		t.Fatalf("expected lay-only definitions to omit extra height, got %#v", bed)
	}
	if bed.UsagePolicy != 0 || bed.OwnerID != 2 {
		t.Fatalf("unexpected bed usage/owner fields %#v", bed)
	}
}

// TestFloorItemsSkipsOrphanedAndUnplacedItems verifies defensive filtering.
func TestFloorItemsSkipsOrphanedAndUnplacedItems(t *testing.T) {
	x, y, z := 1, 1, 0.0
	items := []furnituremodel.Item{
		{DefinitionID: 99, OwnerPlayerID: 1, X: &x, Y: &y, Z: &z},
		{DefinitionID: 2, OwnerPlayerID: 1},
	}
	definitions := map[int64]furnituremodel.Definition{2: {SpriteID: 39}}

	owners, records := FloorItems(items, definitions, nil, nil)
	if len(records) != 0 {
		t.Fatalf("expected no records, got %#v", records)
	}
	if len(owners) != 1 || owners[0].Name != "" {
		t.Fatalf("expected one owner with empty resolved name, got %#v", owners)
	}
}

// TestFloorItemsProjectsGiftWrapper verifies unopened gifts render as presents.
func TestFloorItemsProjectsGiftWrapper(t *testing.T) {
	x, y, z := 1, 2, 0.0
	sprite, box, ribbon := int32(3379), int32(2), int32(7)
	item := furnituremodel.Item{DefinitionID: 2, OwnerPlayerID: 4, X: &x, Y: &y, Z: &z,
		GiftWrapped: true, GiftWrapSpriteID: &sprite, GiftWrapBoxID: &box, GiftWrapRibbonID: &ribbon}
	item.ID = 40
	_, records := FloorItems([]furnituremodel.Item{item}, map[int64]furnituremodel.Definition{2: {SpriteID: 39}}, nil, nil)
	if len(records) != 1 || records[0].SpriteID != 3379 || records[0].Kind != 2007 {
		t.Fatalf("unexpected gift room record %#v", records)
	}
}

// TestExtraHeightValueProjectsTrophyHeight verifies Nitro receives its required trophy height.
func TestExtraHeightValueProjectsTrophyHeight(t *testing.T) {
	definition := furnituremodel.Definition{InteractionType: "trophy", StackHeight: 1}
	if got := ExtraHeightValue(definition); got != "1.0" {
		t.Fatalf("expected trophy extra height 1.0, got %q", got)
	}
}

// TestWallItemsProjectsOnlyPostItColor verifies room entry does not combine note text with sprite state.
func TestWallItemsProjectsOnlyPostItColor(t *testing.T) {
	position := ":w=3,4 l=5,23 r"
	items := []furnituremodel.Item{
		{DefinitionID: 20, OwnerPlayerID: 3, WallPosition: &position, ExtraData: "9CFF9C visible text"},
		{DefinitionID: 21, OwnerPlayerID: 3, WallPosition: &position, ExtraData: "2,1,1,#000000,255"},
	}
	items[0].ID = 1
	items[1].ID = 2
	definitions := map[int64]furnituremodel.Definition{
		20: {Kind: furnituremodel.KindWall, SpriteID: 1, InteractionType: "postit"},
		21: {Kind: furnituremodel.KindWall, SpriteID: 2, InteractionType: "dimmer"},
	}
	_, records := WallItems(items, definitions, map[int64]string{3: "bob"})
	if len(records) != 2 || records[0].ExtraData != "9CFF9C" || records[1].ExtraData != items[1].ExtraData {
		t.Fatalf("unexpected wall extra data %#v", records)
	}
}
