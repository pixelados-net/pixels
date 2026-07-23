package furnitureindex

import (
	"testing"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestBuildAndQueriesVerifyIndexedFurnitureBehavior verifies build, slot, list, and nearest lookups.
func TestBuildAndQueriesVerifyIndexedFurnitureBehavior(t *testing.T) {
	definition := worldfurniture.Definition{
		InteractionType: "pet_food", Width: 1, Length: 1, StackHeight: grid.HeightFromUnits(0.5), AllowWalk: true,
		Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationSouth}},
	}
	items := []worldfurniture.Item{
		{ID: 20, Definition: definition, Point: grid.MustPoint(4, 0)},
		{ID: 10, Definition: definition, Point: grid.MustPoint(1, 0)},
	}
	indexed, err := Build(items)
	if err != nil {
		t.Fatal(err)
	}
	if len(indexed.Fixtures) == 0 || len(indexed.Items) != 2 || len(indexed.Tiles) != 2 || len(indexed.Types["pet_food"]) != 2 || len(indexed.Interactions) != 2 {
		t.Fatalf("indexed=%+v", indexed)
	}
	if goal := ResolveSlotGoal(indexed.Items, grid.MustPoint(1, 0)); goal != grid.MustPoint(1, 0) {
		t.Fatalf("goal=%+v", goal)
	}
	if listed := ByInteraction(indexed.Types["pet_food"], indexed.Items); len(listed) != 2 {
		t.Fatalf("listed=%+v", listed)
	}
	if item, found := Nearest(indexed.Types["pet_food"], indexed.Items, grid.MustPoint(0, 0), 3); !found || item.ID != 10 {
		t.Fatalf("item=%+v found=%v", item, found)
	}
	if _, found := Nearest(indexed.Types["pet_food"], indexed.Items, grid.MustPoint(0, 0), 0); found {
		t.Fatal("expected radius rejection")
	}
}

// TestBuildRejectsInvalidFurniture verifies fixture validation remains visible at room load.
func TestBuildRejectsInvalidFurniture(t *testing.T) {
	_, err := Build([]worldfurniture.Item{{
		ID: 1, Definition: worldfurniture.Definition{Width: 1, Length: 1, AllowWalk: true, StackHeight: -grid.HeightScale},
	}})
	if err == nil {
		t.Fatal("expected invalid furniture")
	}
}

// TestEmptyInteractionQueryReturnsNil verifies absent indexes avoid allocations.
func TestEmptyInteractionQueryReturnsNil(t *testing.T) {
	if items := ByInteraction(nil, nil); items != nil {
		t.Fatalf("items=%+v", items)
	}
}
