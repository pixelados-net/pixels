package teleport

import (
	"context"
	"testing"
	"time"

	teleportpair "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport/pair"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestSameRoomTransitionOpensCrossesAndSettles verifies the complete pad cycle.
func TestSameRoomTransitionOpensCrossesAndSettles(t *testing.T) {
	service, active, now := serviceForTest(t, "teleport")
	if err := service.Start(context.Background(), StartRequest{PlayerID: 7, Room: active, ItemID: 1}); err != nil {
		t.Fatalf("start teleport: %v", err)
	}
	if err := service.Cycle(context.Background(), active, now); err != nil {
		t.Fatalf("open source: %v", err)
	}
	if item, _ := active.FurnitureItem(1); item.ExtraData != "1" {
		t.Fatalf("expected open source, got %q", item.ExtraData)
	}
	unit, _ := active.Unit(7)
	if unit.Position.Point != grid.MustPoint(1, 2) || !unit.Moving {
		t.Fatalf("expected animated source entry, got %#v", unit)
	}
	active.Tick()
	if err := service.Cycle(context.Background(), active, now); err != nil {
		t.Fatalf("wait for visible source entry: %v", err)
	}
	active.Tick()
	if err := service.Cycle(context.Background(), active, now.Add(phaseDelay)); err != nil {
		t.Fatalf("settle source entry: %v", err)
	}
	if err := service.Cycle(context.Background(), active, now.Add(2*phaseDelay)); err != nil {
		t.Fatalf("cross teleport: %v", err)
	}
	unit, _ = active.Unit(7)
	if unit.Position.Point != grid.MustPoint(3, 1) {
		t.Fatalf("expected target position, got %#v", unit.Position)
	}
	if err := service.Cycle(context.Background(), active, now.Add(3*phaseDelay)); err != nil {
		t.Fatalf("start exit: %v", err)
	}
	active.Tick()
	if err := service.Cycle(context.Background(), active, now.Add(4*phaseDelay)); err != nil {
		t.Fatalf("animate exit: %v", err)
	}
	active.Tick()
	if err := service.Cycle(context.Background(), active, now.Add(5*phaseDelay)); err != nil {
		t.Fatalf("settle exit: %v", err)
	}
	unit, _ = active.Unit(7)
	if unit.Position.Point != grid.MustPoint(3, 2) || unit.Moving {
		t.Fatalf("expected settled exit, got %#v", unit)
	}
	if item, _ := active.FurnitureItem(2); item.ExtraData != "0" {
		t.Fatalf("expected closed target, got %q", item.ExtraData)
	}
}

// TestTileTransitionAdvancesWithoutPhaseDelay verifies the tile variant in one cycle.
func TestTileTransitionAdvancesWithoutPhaseDelay(t *testing.T) {
	service, active, now := serviceForTest(t, "teleport_tile")
	if err := service.Start(context.Background(), StartRequest{PlayerID: 7, Room: active, ItemID: 1}); err != nil {
		t.Fatalf("start tile teleport: %v", err)
	}
	if err := service.Cycle(context.Background(), active, now); err != nil {
		t.Fatalf("cycle tile teleport: %v", err)
	}
	active.Tick()
	if err := service.Cycle(context.Background(), active, now); err != nil {
		t.Fatalf("animate tile teleport entry: %v", err)
	}
	active.Tick()
	if err := service.Cycle(context.Background(), active, now); err != nil {
		t.Fatalf("complete tile teleport entry: %v", err)
	}
	unit, _ := active.Unit(7)
	if unit.Position.Point != grid.MustPoint(3, 1) && unit.Position.Point != grid.MustPoint(3, 2) {
		t.Fatalf("expected target-side tile, got %#v", unit.Position)
	}
}

// TestStartApproachesFromDistanceAndIgnoresDuplicate verifies controlled setup behavior.
func TestStartApproachesFromDistanceAndIgnoresDuplicate(t *testing.T) {
	service, active, _ := serviceForTest(t, "teleport")
	if _, err := active.TeleportUnit(7, grid.MustPoint(0, 2), worldunit.RotationEast, false); err != nil {
		t.Fatalf("reposition unit: %v", err)
	}
	request := StartRequest{PlayerID: 7, Room: active, ItemID: 1}
	if err := service.Start(context.Background(), request); err != nil {
		t.Fatalf("start distant teleport: %v", err)
	}
	unit, _ := active.Unit(7)
	if !unit.Moving {
		t.Fatal("expected approach path")
	}
	if err := service.Start(context.Background(), request); err != nil {
		t.Fatalf("duplicate use should be ignored: %v", err)
	}
}

// TestCycleHandlesRemovedSource verifies an animation abort is soft.
func TestCycleHandlesRemovedSource(t *testing.T) {
	service, active, now := serviceForTest(t, "teleport")
	if err := service.Start(context.Background(), StartRequest{PlayerID: 7, Room: active, ItemID: 1}); err != nil {
		t.Fatalf("start teleport: %v", err)
	}
	if _, err := active.ReloadFurniture(1, nil); err != nil {
		t.Fatalf("remove source: %v", err)
	}
	if err := service.Cycle(context.Background(), active, now); err != nil {
		t.Fatalf("abort removed source: %v", err)
	}
}

// TestStartWithoutPairReleasesReservation verifies unpaired furniture remains inert.
func TestStartWithoutPairReleasesReservation(t *testing.T) {
	service, active, _ := serviceForTest(t, "teleport")
	service.pairs = teleportpair.NewService(emptyPairStoreForTest{}, &teleportFurnitureForTest{})
	if err := service.Start(context.Background(), StartRequest{PlayerID: 7, Room: active, ItemID: 1}); err != nil {
		t.Fatalf("unpaired teleport should be inert: %v", err)
	}
	if _, found := service.rooms.Load(active.ID()); found {
		t.Fatal("expected unpaired reservation cleanup")
	}
}

// BenchmarkCycleWithoutTransitions measures the dominant active-room cycle path.
func BenchmarkCycleWithoutTransitions(b *testing.B) {
	service, active, now := serviceForTest(b, "teleport")
	b.ReportAllocs()
	for b.Loop() {
		_ = service.Cycle(context.Background(), active, now)
	}
}

// serviceForTest creates a loaded room and paired teleport service.
func serviceForTest(t testing.TB, interactionType string) (*Service, *roomlive.Room, time.Time) {
	t.Helper()
	now := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	definition := worldfurniture.Definition{SpriteID: 202, InteractionType: interactionType, Width: 1, Length: 1}
	items := []worldfurniture.Item{
		{ID: 1, OwnerPlayerID: 7, Definition: definition, Point: grid.MustPoint(1, 1), Rotation: worldunit.RotationSouth, ExtraData: "0"},
		{ID: 2, OwnerPlayerID: 7, Definition: definition, Point: grid.MustPoint(3, 1), Rotation: worldunit.RotationSouth, ExtraData: "0"},
	}
	roomGrid, err := grid.Parse("00000\r00000\r00000", grid.WithDoor(1, 2))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if err := active.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid, Furniture: items, Door: worldpath.Position{Point: grid.MustPoint(1, 2)},
		Body: worldunit.RotationNorth, Head: worldunit.RotationNorth,
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, roomlive.Occupant{
		PlayerID: 7, Username: "demo", ConnectionID: "one", ConnectionKind: "websocket",
	}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	roomID, firstX, firstY, secondX, secondY, z := int64(9), 1, 1, 3, 1, 0.0
	furniture := &teleportFurnitureForTest{
		items: map[int64]furnituremodel.Item{
			1: {Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 8, OwnerPlayerID: 7, RoomID: &roomID, X: &firstX, Y: &firstY, Z: &z},
			2: {Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, DefinitionID: 8, OwnerPlayerID: 7, RoomID: &roomID, X: &secondX, Y: &secondY, Z: &z},
		},
		definition: furnituremodel.Definition{SpriteID: 202, Width: 1, Length: 1, InteractionType: interactionType},
	}
	pairs := teleportpair.NewService(teleportPairStoreForTest{}, furniture)
	service := NewService(Config{}, pairs, runtime, netconn.NewRegistry(), nil, bus.New())
	service.now = func() time.Time { return now }

	return service, active, now
}

// teleportPairStoreForTest resolves one fixed pair.
type teleportPairStoreForTest struct{}

// emptyPairStoreForTest reports no relationship.
type emptyPairStoreForTest struct{}

// FindByItem reports no pair.
func (emptyPairStoreForTest) FindByItem(context.Context, int64) (teleportpair.Pair, bool, error) {
	return teleportpair.Pair{}, false, nil
}

// Replace stores no pair.
func (emptyPairStoreForTest) Replace(context.Context, teleportpair.Pair) error { return nil }

// DeleteByItem removes no pair.
func (emptyPairStoreForTest) DeleteByItem(context.Context, int64) (bool, error) { return false, nil }

// FindByItem finds the fixed pair.
func (teleportPairStoreForTest) FindByItem(context.Context, int64) (teleportpair.Pair, bool, error) {
	return teleportpair.Pair{ItemOneID: 1, ItemTwoID: 2}, true, nil
}

// Replace stores no test state.
func (teleportPairStoreForTest) Replace(context.Context, teleportpair.Pair) error { return nil }

// DeleteByItem removes no test state.
func (teleportPairStoreForTest) DeleteByItem(context.Context, int64) (bool, error) { return true, nil }

// teleportFurnitureForTest stores target fixtures.
type teleportFurnitureForTest struct {
	// items stores item fixtures.
	items map[int64]furnituremodel.Item
	// definition stores the shared definition fixture.
	definition furnituremodel.Definition
}

// FindItemByID finds an item fixture.
func (store *teleportFurnitureForTest) FindItemByID(_ context.Context, id int64) (furnituremodel.Item, bool, error) {
	item, found := store.items[id]
	return item, found, nil
}

// FindDefinitionByID returns the shared definition fixture.
func (store *teleportFurnitureForTest) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return store.definition, true, nil
}
