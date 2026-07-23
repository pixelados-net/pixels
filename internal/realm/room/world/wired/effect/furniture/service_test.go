package furniture

import (
	"context"
	"testing"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// furnitureManager stores one durable item and definition for effect integration tests.
type furnitureManager struct {
	// item stores the mutable durable item.
	item furnituremodel.Item
	// definition stores the item's immutable definition.
	definition furnituremodel.Definition
}

// FindDefinitionByID returns the configured definition.
func (manager *furnitureManager) FindDefinitionByID(_ context.Context, id int64) (furnituremodel.Definition, bool, error) {
	return manager.definition, id == manager.definition.ID, nil
}

// ListDefinitions returns the configured definition.
func (manager *furnitureManager) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return []furnituremodel.Definition{manager.definition}, nil
}

// FindItemByID returns the configured item.
func (manager *furnitureManager) FindItemByID(_ context.Context, id int64) (furnituremodel.Item, bool, error) {
	return manager.item, id == manager.item.ID, nil
}

// ListInventory is unused by effect integration tests.
func (*furnitureManager) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}

// ListRoomItems returns the configured item.
func (manager *furnitureManager) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return []furnituremodel.Item{manager.item}, nil
}

// Place is unused by effect integration tests.
func (manager *furnitureManager) Place(context.Context, furnitureservice.PlaceParams) (furnituremodel.Item, error) {
	return manager.item, nil
}

// Move updates the configured durable placement.
func (manager *furnitureManager) Move(_ context.Context, params furnitureservice.MoveParams) (furnituremodel.Item, error) {
	manager.item.X, manager.item.Y, manager.item.Z = intPointer(params.Placement.X), intPointer(params.Placement.Y), floatPointer(params.Placement.Z)
	manager.item.Rotation = params.Placement.Rotation
	return manager.item, nil
}

// Pickup is unused by effect integration tests.
func (manager *furnitureManager) Pickup(context.Context, furnitureservice.PickupParams) (furnituremodel.Item, error) {
	return manager.item, nil
}

// UpdateState updates the configured durable state.
func (manager *furnitureManager) UpdateState(_ context.Context, params furnitureservice.StateParams) (furnituremodel.Item, error) {
	manager.item.ExtraData = params.Next
	return manager.item, nil
}

// TestMovementHelpersRemainDeterministic verifies stable geometry ordering used by movement effects.
func TestMovementHelpersRemainDeterministic(t *testing.T) {
	points := []grid.Point{grid.MustPoint(1, 0), grid.MustPoint(0, 1), grid.MustPoint(2, 2)}
	sortByDistance(points, grid.MustPoint(0, 0), false)
	if points[0] != grid.MustPoint(1, 0) || points[2] != grid.MustPoint(2, 2) {
		t.Fatalf("toward order=%v", points)
	}
	sortByDistance(points, grid.MustPoint(0, 0), true)
	if points[0] != grid.MustPoint(2, 2) {
		t.Fatalf("away order=%v", points)
	}
	if distance(grid.MustPoint(0, 0), grid.MustPoint(2, 2)) != 4 {
		t.Fatal("invalid Manhattan distance")
	}
	if len(cardinal(grid.MustPoint(0, 0))) != 2 {
		t.Fatal("origin cardinal points must remain in bounds")
	}
	node := &configuration.Node{Parameters: configuration.Parameters{Values: []int32{2}}}
	if movementMode(node) != 2 || movementMode(&configuration.Node{}) != 0 {
		t.Fatal("movement mode parsing failed")
	}
	if durableRotation(worldunit.RotationEast) != 2 {
		t.Fatal("durable rotation mapping failed")
	}
}

// TestServiceFailsClosedWithoutTargets verifies missing rooms, furniture, and managers are no-ops.
func TestServiceFailsClosedWithoutTargets(t *testing.T) {
	rooms := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	service := New(rooms, nil, nil)
	result, err := service.ExecuteFurniture(context.Background(), effect.ToggleState, &configuration.Node{}, trigger.Event{RoomID: 99})
	if err != nil || result.Status != effect.Skipped {
		t.Fatalf("missing room result=%+v err=%v", result, err)
	}
	active, err := rooms.Activate(roomlive.Snapshot{ID: 99, MaxUsers: 2})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = rooms.Close(context.Background(), active.ID()) })
	node := &configuration.Node{Targets: []record.Target{{ItemID: 123}}}
	result, err = service.ExecuteFurniture(context.Background(), effect.ToggleState, node, trigger.Event{RoomID: 99})
	if err != nil || result.Status != effect.Skipped {
		t.Fatalf("missing target result=%+v err=%v", result, err)
	}
	if err = service.Activate(context.Background(), 99, 123); err != nil {
		t.Fatal(err)
	}
	if !containsTarget(node.Targets, 123) || containsTarget(node.Targets, 124) {
		t.Fatal("target membership failed")
	}
}

// TestServiceMutatesFurnitureThroughAuthoritativePlacement verifies toggle, movement, and snapshot restoration.
func TestServiceMutatesFurnitureThroughAuthoritativePlacement(t *testing.T) {
	rooms, manager, active := furnitureRoom(t)
	service := New(rooms, manager, nil)
	event := trigger.Event{ID: 3, RoomID: active.ID(), PlayerID: 7, ActorID: 7, ActorKind: trigger.ActorPlayer}
	node := &configuration.Node{Targets: []record.Target{{ItemID: 10}}}
	result, err := service.ExecuteFurniture(context.Background(), effect.ToggleState, node, event)
	if err != nil || result.Status != effect.Applied || len(result.Derived) != 1 {
		t.Fatalf("toggle result=%+v err=%v", result, err)
	}
	item, _ := active.FurnitureItem(10)
	if item.ExtraData != "1" {
		t.Fatalf("toggle state=%q", item.ExtraData)
	}
	result, err = service.ExecuteFurniture(context.Background(), effect.ToggleRandomState, node, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("random toggle result=%+v err=%v", result, err)
	}
	node.Parameters.Values = []int32{2}
	result, err = service.ExecuteFurniture(context.Background(), effect.MoveDirection, node, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("move result=%+v err=%v", result, err)
	}
	item, _ = active.FurnitureItem(10)
	if item.Point != grid.MustPoint(2, 0) {
		t.Fatalf("moved point=%+v", item.Point)
	}
	node.Targets[0].Snapshot = record.Snapshot{State: "0", X: 1, Y: 0, Z: 0, Rotation: 0, Present: true}
	result, err = service.ExecuteFurniture(context.Background(), effect.MatchSnapshot, node, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("restore result=%+v err=%v", result, err)
	}
	item, _ = active.FurnitureItem(10)
	if item.Point != grid.MustPoint(1, 0) || item.ExtraData != "0" {
		t.Fatalf("restored item=%+v", item)
	}
	if err = service.Activate(context.Background(), active.ID(), 10); err != nil {
		t.Fatal(err)
	}
}

// TestFleeCollisionRepeatsForEveryTouchAttempt verifies collision is neither a user-path miss nor a one-shot trigger.
func TestFleeCollisionRepeatsForEveryTouchAttempt(t *testing.T) {
	rooms, manager, active := furnitureRoom(t)
	if _, err := active.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	service := New(rooms, manager, nil)
	event := trigger.Event{ID: 1, RoomID: active.ID(), PlayerID: 8, ActorID: 8, ActorKind: trigger.ActorPlayer}
	node := &configuration.Node{Targets: []record.Target{{ItemID: 10}}}
	for attempt := 1; attempt <= 2; attempt++ {
		event.ID = uint64(attempt)
		result, err := service.ExecuteFurniture(context.Background(), effect.FleeActor, node, event)
		if err != nil || result.Status != effect.Applied || len(result.Derived) != 1 {
			t.Fatalf("attempt %d result=%+v err=%v", attempt, result, err)
		}
		collision := result.Derived[0]
		if collision.Kind != trigger.Collision || collision.ActorID != 7 || collision.PlayerID != 7 || collision.SourceItem != 10 {
			t.Fatalf("attempt %d collision=%+v", attempt, collision)
		}
	}
	item, _ := active.FurnitureItem(10)
	if item.Point != grid.MustPoint(1, 0) {
		t.Fatalf("collided item moved to %+v", item.Point)
	}
}

// furnitureRoom creates one movable multi-state furniture item in an active room.
func furnitureRoom(t *testing.T) (*roomlive.Registry, *furnitureManager, *roomlive.Room) {
	t.Helper()
	roomID, x, y, z := int64(55), 1, 0, 0.0
	definition := furnituremodel.Definition{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, SpriteID: 2, Kind: furnituremodel.KindFloor, Width: 1, Length: 1, AllowWalk: true, InteractionModesCount: 3}
	durable := furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 10}}, DefinitionID: 1, OwnerPlayerID: 7, RoomID: &roomID, X: &x, Y: &y, Z: &z, ExtraData: "0"}
	manager := &furnitureManager{item: durable, definition: definition}
	rooms := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := rooms.Activate(roomlive.Snapshot{ID: roomID, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	worldItem := worldfurniture.Item{ID: 10, OwnerPlayerID: 7, Point: grid.MustPoint(1, 0), ExtraData: "0", Definition: worldfurniture.Definition{SpriteID: 2, Width: 1, Length: 1, AllowWalk: true, InteractionModesCount: 3}}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{worldItem}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = rooms.Close(context.Background(), roomID) })
	return rooms, manager, active
}

// intPointer returns a pointer to one test integer.
func intPointer(value int) *int { return &value }

// floatPointer returns a pointer to one test float.
func floatPointer(value float64) *float64 { return &value }
