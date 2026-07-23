package roller

import (
	"context"
	"testing"
	"time"

	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestCycleMovesStackOneRollerPerTick verifies chain deduplication and relative stack heights.
func TestCycleMovesStackOneRollerPerTick(t *testing.T) {
	items := []worldfurniture.Item{rollerItem(1, 1), rollerItem(2, 2), rollerItem(3, 3), stackedItem(10, 1, grid.Height(2)), stackedItem(11, 1, grid.Height(6))}
	active := rollerRoom(t, items)
	service := testService(nil)
	if err := service.Cycle(context.Background(), active, time.Now()); err != nil {
		t.Fatalf("first cycle: %v", err)
	}
	first, _ := active.FurnitureItem(10)
	second, _ := active.FurnitureItem(11)
	if first.Point != grid.MustPoint(2, 0) || second.Point != grid.MustPoint(2, 0) || second.Z-first.Z != 4 {
		t.Fatalf("unexpected first roll first=%#v second=%#v", first, second)
	}
	if err := service.Cycle(context.Background(), active, time.Now()); err != nil {
		t.Fatalf("second cycle: %v", err)
	}
	first, _ = active.FurnitureItem(10)
	if first.Point != grid.MustPoint(3, 0) {
		t.Fatalf("expected one additional tile, got %#v", first.Point)
	}
}

// TestCycleSkipsWalkingUnit verifies player movement wins over a roller cycle.
func TestCycleSkipsWalkingUnit(t *testing.T) {
	active := rollerRoom(t, []worldfurniture.Item{rollerItem(1, 1)})
	joinAndPlace(t, active, 7, grid.MustPoint(1, 0))
	if _, err := active.MoveTo(7, grid.MustPoint(2, 0)); err != nil {
		t.Fatalf("start walk: %v", err)
	}
	if err := testService(nil).Cycle(context.Background(), active, time.Now()); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	unit, _ := active.Unit(7)
	if unit.Position.Point != grid.MustPoint(1, 0) || !unit.Moving {
		t.Fatalf("walking unit was rolled %#v", unit)
	}
}

// TestCycleTargetUnitBlocksRoller verifies any destination unit blocks the complete step.
func TestCycleTargetUnitBlocksRoller(t *testing.T) {
	active := rollerRoom(t, []worldfurniture.Item{rollerItem(1, 1), stackedItem(10, 1, 2)})
	joinAndPlace(t, active, 8, grid.MustPoint(2, 0))
	if err := testService(nil).Cycle(context.Background(), active, time.Now()); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	item, _ := active.FurnitureItem(10)
	if item.Point != grid.MustPoint(1, 0) {
		t.Fatalf("blocked item moved to %#v", item.Point)
	}
}

// TestCycleDoesNotMoveFurnitureUphill verifies furniture cannot climb above the belt surface.
func TestCycleDoesNotMoveFurnitureUphill(t *testing.T) {
	platform := stackedItem(20, 2, 0)
	platform.Definition.StackHeight = grid.HeightFromInt(1)
	platform.Definition.AllowStack = true
	active := rollerRoom(t, []worldfurniture.Item{rollerItem(1, 1), stackedItem(10, 1, 2), platform})
	if err := testService(nil).Cycle(context.Background(), active, time.Now()); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	item, _ := active.FurnitureItem(10)
	if item.Point != grid.MustPoint(1, 0) {
		t.Fatalf("uphill furniture moved to %#v", item.Point)
	}
}

// TestCycleDelaysWalkHooks verifies hooks run only after the configured animation delay.
func TestCycleDelaysWalkHooks(t *testing.T) {
	local := bus.New()
	count := 0
	_, err := local.Subscribe(furniturewalkedoff.Name, bus.PriorityNormal, func(context.Context, bus.Event) error { count++; return nil })
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	active := rollerRoom(t, []worldfurniture.Item{rollerItem(1, 1)})
	joinAndPlace(t, active, 7, grid.MustPoint(1, 0))
	service := testService(local)
	if err = service.Cycle(context.Background(), active, time.Now()); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	if count != 0 {
		t.Fatalf("hook ran immediately count=%d", count)
	}
	active.RunScheduled(time.Now().Add(time.Second))
	if count != 1 {
		t.Fatalf("expected delayed hook count=1, got %d", count)
	}
}

// BenchmarkRollerTick measures an indexed chain with mounted furniture.
func BenchmarkRollerTick(b *testing.B) {
	first, second := rollerItem(1, 1), rollerItem(2, 2)
	second.Rotation = worldunit.RotationWest
	active := rollerRoom(b, []worldfurniture.Item{first, second, stackedItem(10, 1, 2)})
	service := testService(nil)
	b.ResetTimer()
	for range b.N {
		if err := service.Cycle(context.Background(), active, time.Now()); err != nil {
			b.Fatal(err)
		}
	}
}

// rollerRoom creates one loaded fixed-point test room.
func rollerRoom(testingT testing.TB, items []worldfurniture.Item) *roomlive.Room {
	testingT.Helper()
	roomGrid, err := grid.Parse("00000", grid.WithDoor(0, 0))
	if err != nil {
		testingT.Fatal(err)
	}
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 25, RollerSpeed: 0})
	if err != nil {
		testingT.Fatal(err)
	}
	err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Furniture: items})
	if err != nil {
		testingT.Fatal(err)
	}
	return active
}

// rollerItem creates an east-facing half-unit roller.
func rollerItem(id int64, x int) worldfurniture.Item {
	return worldfurniture.Item{ID: id, OwnerPlayerID: 7, Point: grid.MustPoint(x, 0), Rotation: worldunit.RotationEast,
		Definition: worldfurniture.Definition{SpriteID: 7, InteractionType: "roller", Width: 1, Length: 1, StackHeight: 2, AllowStack: true, AllowWalk: true}}
}

// stackedItem creates one stackable mounted test item.
func stackedItem(id int64, x int, z grid.Height) worldfurniture.Item {
	return worldfurniture.Item{ID: id, OwnerPlayerID: 7, Point: grid.MustPoint(x, 0), Z: z,
		Definition: worldfurniture.Definition{SpriteID: int(id), Width: 1, Length: 1, StackHeight: 4, AllowStack: true}}
}

// joinAndPlace joins and directly settles one player on a tile.
func joinAndPlace(t *testing.T, active *roomlive.Room, playerID int64, point grid.Point) {
	t.Helper()
	_, err := active.Join(roomlive.Occupant{PlayerID: playerID, ConnectionID: "test", ConnectionKind: "websocket"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = active.TeleportUnit(playerID, point, worldunit.RotationSouth, false); err != nil {
		t.Fatal(err)
	}
}

// testService creates roller behavior without starting persistence.
func testService(events *bus.Bus) *Service {
	service := &Service{config: Config{Delay: 400 * time.Millisecond, MaxAvatarsPerTick: 1}.Normalize(), persistence: make(chan persistence, 32)}
	if events != nil {
		service.events = events
	}
	return service
}
