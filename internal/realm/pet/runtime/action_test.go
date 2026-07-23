package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// actionClock returns one fixed wall time.
type actionClock struct {
	// now stores the configured wall time.
	now time.Time
}

// Now returns the configured wall time.
func (clock actionClock) Now() time.Time { return clock.now }

// actionStore returns a deterministic optimistic update decision.
type actionStore struct {
	petrecord.Store
	// updated stores whether the optimistic write succeeded.
	updated bool
	// energy stores the requested energy delta.
	energy int32
	// happiness stores the requested happiness delta.
	happiness int32
	// calls counts requested stat mutations.
	calls int
}

// UpdateStats returns the configured optimistic decision.
func (store *actionStore) UpdateStats(_ context.Context, _ int64, energy int32, happiness int32, _ int32, _ int64) (petrecord.Pet, bool, error) {
	store.calls++
	store.energy, store.happiness = energy, happiness
	return petrecord.Pet{}, store.updated, nil
}

// TestFindByUnitUsesReverseIndex verifies hand-item targeting resolves a pet.
func TestFindByUnitUsesReverseIndex(t *testing.T) {
	service, active, unitID := unitLookupFixture(t)
	pet, found := service.FindByUnit(9, unitID)
	if !found || pet.ID != 50 {
		t.Fatalf("pet=%+v found=%v", pet, found)
	}
	if _, found = service.FindByUnit(9, unitID+1); found {
		t.Fatal("expected unknown unit rejection")
	}
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
}

// TestFindByUnitAllocatesNothing verifies the reverse lookup hot path.
func TestFindByUnitAllocatesNothing(t *testing.T) {
	service, active, unitID := unitLookupFixture(t)
	allocations := testing.AllocsPerRun(1000, func() {
		_, _ = service.FindByUnit(9, unitID)
	})
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
	if allocations != 0 {
		t.Fatalf("expected zero allocations, got %.2f", allocations)
	}
}

// TestExecuteActionReportsOptimisticConflict verifies failed stat writes are not acknowledged.
func TestExecuteActionReportsOptimisticConflict(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	store := &actionStore{}
	service.store = store
	err := service.ExecuteAction(context.Background(), active.ID(), 50, 1, CommandAction{ID: 1, Mode: ActionStatus, StatusKey: "sit"}, petrecord.Command{RequiredLevel: 1, EnergyCost: 1, HappinessCost: 2})
	unit, found := active.Unit(EntityKey(50))
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
	if !errors.Is(err, petrecord.ErrConflict) {
		t.Fatalf("expected optimistic conflict, got %v", err)
	}
	if !found || unitHasStatus(unit, "sit") {
		t.Fatalf("failed transaction changed action state %+v", unit)
	}
	if store.energy != -1 || store.happiness != -2 {
		t.Fatalf("unexpected stat deltas energy=%d happiness=%d", store.energy, store.happiness)
	}
}

// TestExecuteActionStopsMovementBeforeStatus verifies a trick replaces autonomous movement immediately.
func TestExecuteActionStopsMovementBeforeStatus(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	if _, err := active.MoveTo(EntityKey(50), grid.MustPoint(0, 0)); err != nil {
		t.Fatal(err)
	}
	before, found := active.Unit(EntityKey(50))
	if !found || !before.Moving {
		t.Fatalf("expected moving pet %+v", before)
	}
	err := service.ExecuteAction(context.Background(), active.ID(), 50, 1, CommandAction{ID: 2, Mode: ActionStatus, StatusKey: worldunit.StatusLay}, petrecord.Command{RequiredLevel: 1})
	after, found := active.Unit(EntityKey(50))
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
	if err != nil {
		t.Fatal(err)
	}
	if !found || after.Moving || !unitHasStatus(after, worldunit.StatusLay) || unitHasStatus(after, worldunit.StatusMove) {
		t.Fatalf("expected stopped lay status %+v", after)
	}
}

// TestDismountPlayerIgnoresLoadingGeneration verifies disconnect cannot dereference a nil snapshot.
func TestDismountPlayerIgnoresLoadingGeneration(t *testing.T) {
	service := &Service{active: map[int64]*roomState{9: {ready: make(chan struct{})}}}
	service.DismountPlayer(context.Background(), 7)
}

// TestMountSynchronizesAndDisconnectCleansStatus verifies rider lifecycle on the shared room world.
func TestMountSynchronizesAndDisconnectCleansStatus(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	controller, found := service.Active(active.ID(), 50)
	if !found {
		t.Fatal("expected pet controller")
	}
	controller.mutex.Lock()
	controller.record.HasSaddle = true
	controller.record.PublicRide = true
	controller.mutex.Unlock()
	if _, err := active.Join(roomlive.Occupant{PlayerID: 2, Username: "rider", ConnectionID: netconn.ID("test"), ConnectionKind: netconn.Kind("ws")}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Mount(context.Background(), active.ID(), 50, 2, true); err != nil {
		t.Fatal(err)
	}
	if rider, riding := service.Rider(active.ID(), 50); !riding || rider != 2 {
		t.Fatalf("rider=%d riding=%v", rider, riding)
	}
	mountedPlayer, found := active.Unit(2)
	if !found || mountedPlayer.RenderOffset != worldunit.RidingHeightOffset || unitHasStatus(mountedPlayer, worldunit.StatusSit) {
		t.Fatalf("mounted player=%+v found=%v", mountedPlayer, found)
	}
	if _, err := active.TeleportUnit(2, grid.MustPoint(0, 0), worldunit.RotationNorth, false); err != nil {
		t.Fatal(err)
	}
	petUnit, found := active.Unit(EntityKey(50))
	if !found || petUnit.Position.Point != grid.MustPoint(0, 0) {
		t.Fatalf("pet unit=%+v found=%v", petUnit, found)
	}
	service.DismountPlayer(context.Background(), 2)
	if _, riding := service.Rider(active.ID(), 50); riding {
		t.Fatal("expected rider cleanup")
	}
	playerUnit, found := active.Unit(2)
	if !found || playerUnit.RenderOffset != 0 || unitHasStatus(playerUnit, worldunit.StatusSit) {
		t.Fatalf("player unit=%+v found=%v", playerUnit, found)
	}
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
}

// TestNowUsesInjectedClock verifies every pet lifecycle can share deterministic time.
func TestNowUsesInjectedClock(t *testing.T) {
	expected := time.Date(2026, time.July, 17, 12, 0, 0, 0, time.UTC)
	service := &Service{}
	service.SetClock(actionClock{now: expected})
	if actual := service.Now(); !actual.Equal(expected) {
		t.Fatalf("now=%s", actual)
	}
}

// TestTrainingCommandsFiltersLevelAndDisabled verifies learned command projection.
func TestTrainingCommandsFiltersLevelAndDisabled(t *testing.T) {
	references := trainingFixture()
	all, enabled := petreference.TrainingCommands(references, petrecord.Pet{TypeID: 0, Level: 2})
	if len(all) != 3 || len(enabled) != 1 || enabled[0] != 1 {
		t.Fatalf("all=%v enabled=%v", all, enabled)
	}
}

// BenchmarkPetUnitLookup measures room-local unit-to-pet resolution.
func BenchmarkPetUnitLookup(b *testing.B) {
	service, active, unitID := unitLookupFixture(b)
	b.Cleanup(func() { _, _, _ = service.rooms.Close(context.Background(), active.ID()) })
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, _ = service.FindByUnit(9, unitID)
	}
}

// BenchmarkTrainingSnapshot measures immutable learned-command selection.
func BenchmarkTrainingSnapshot(b *testing.B) {
	references := trainingFixture()
	pet := petrecord.Pet{TypeID: 0, Level: 2}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		_, _ = petreference.TrainingCommands(references, pet)
	}
}

// unitLookupFixture creates one loaded room with one indexed pet controller.
func unitLookupFixture(t testing.TB) (*Service, *roomlive.Room, int64) {
	t.Helper()
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 25, AllowPets: true, AllowPetsEat: true})
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	unit, err := active.AddEntity(EntityKey(50), 1, worldunit.KindPet, worldpath.Position{Point: grid.MustPoint(1, 0)}, worldunit.RotationSouth)
	if err != nil {
		t.Fatal(err)
	}
	pet := &activePet{record: petrecord.Pet{ID: 50, OwnerPlayerID: 1, Level: 1, Energy: 100, Happiness: 100, Version: 1}}
	state := &roomState{pets: map[int64]*activePet{50: pet}}
	state.rebuildSnapshot()
	return &Service{rooms: rooms, active: map[int64]*roomState{9: state}}, active, unit.UnitID
}

// unitHasStatus reports whether one unit snapshot contains a status key.
func unitHasStatus(unit roomlive.UnitSnapshot, key string) bool {
	for _, status := range unit.Statuses {
		if status.Key == key {
			return true
		}
	}
	return false
}

// trainingFixture creates one immutable command generation.
func trainingFixture() *petreference.Snapshot {
	references := &petreference.Snapshot{}
	references.SpeciesCommands[0] = []int32{1, 2, 3}
	references.Commands[1] = petrecord.Command{ID: 1, RequiredLevel: 2, Enabled: true}
	references.Commands[2] = petrecord.Command{ID: 2, RequiredLevel: 3, Enabled: true}
	references.Commands[3] = petrecord.Command{ID: 3, RequiredLevel: 1, Enabled: false}
	references.CommandPresent[1], references.CommandPresent[2], references.CommandPresent[3] = true, true, true
	return references
}
