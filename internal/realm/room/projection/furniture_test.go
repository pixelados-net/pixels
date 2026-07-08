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

	owners, records := FloorItems(items, definitions, ownerNames)

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

	owners, records := FloorItems(items, definitions, nil)
	if len(records) != 0 {
		t.Fatalf("expected no records, got %#v", records)
	}
	if len(owners) != 1 || owners[0].Name != "" {
		t.Fatalf("expected one owner with empty resolved name, got %#v", owners)
	}
}
