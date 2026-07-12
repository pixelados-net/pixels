package furniture

import (
	"context"
	"errors"
	"testing"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestResolveWorldItemBuildsResolvedItem verifies successful target resolution.
func TestResolveWorldItemBuildsResolvedItem(t *testing.T) {
	active := roomForPlacementTest(t)
	manager := &fakeManager{definitions: []furnituremodel.Definition{chairDefinitionForTest()}}

	persisted := furnituremodel.Item{DefinitionID: 2, OwnerPlayerID: 9, ExtraData: "3"}
	persisted.ID = 7
	item, definition, err := ResolveWorldItem(context.Background(), active, manager, persisted, 3, 0, furnituremodel.RotationNorth)
	if err != nil {
		t.Fatalf("resolve world item: %v", err)
	}
	if item.ID != 7 || item.Point.X != 3 || item.Point.Y != 0 || item.Z != 0 {
		t.Fatalf("unexpected resolved item %#v", item)
	}
	if definition.ID != 2 {
		t.Fatalf("unexpected resolved definition %#v", definition)
	}
	if item.OwnerPlayerID != 9 || item.ExtraData != "3" {
		t.Fatalf("placement state was not preserved: %#v", item)
	}
}

// TestResolveWorldItemRejectsInvalidRotation verifies rotation validation.
func TestResolveWorldItemRejectsInvalidRotation(t *testing.T) {
	active := roomForPlacementTest(t)
	manager := &fakeManager{definitions: []furnituremodel.Definition{chairDefinitionForTest()}}

	item := furnituremodel.Item{DefinitionID: 2}
	item.ID = 7
	_, _, err := ResolveWorldItem(context.Background(), active, manager, item, 3, 3, furnituremodel.Rotation(1))
	if !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected invalid target, got %v", err)
	}
}

// TestResolveWorldItemRejectsMissingDefinition verifies the definition-lookup guard.
func TestResolveWorldItemRejectsMissingDefinition(t *testing.T) {
	active := roomForPlacementTest(t)
	manager := &fakeManager{}

	item := furnituremodel.Item{DefinitionID: 2}
	item.ID = 7
	_, _, err := ResolveWorldItem(context.Background(), active, manager, item, 3, 3, furnituremodel.RotationNorth)
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("expected definition not found, got %v", err)
	}
}

// roomForPlacementTest creates a loaded flat room for placement tests.
func roomForPlacementTest(t *testing.T) *roomlive.Room {
	t.Helper()

	room, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 25})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	roomGrid, err := grid.Parse("0000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	if err := room.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid,
		Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
		Body: worldunit.RotationSouth,
		Head: worldunit.RotationSouth,
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}

	return room
}
