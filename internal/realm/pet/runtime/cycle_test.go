package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// fixedSource returns one deterministic runtime value.
type fixedSource struct {
	// value stores the repeated pseudo-random value.
	value uint64
}

// Uint64 returns the configured deterministic value.
func (source fixedSource) Uint64() uint64 { return source.value }

// TestDecisionDelayStaysBounded verifies autonomous jitter boundaries.
func TestDecisionDelayStaysBounded(t *testing.T) {
	service := &Service{config: petpolicy.Config{DecisionMinimum: 2 * time.Second, DecisionMaximum: 4 * time.Second}, source: fixedSource{value: uint64(3 * time.Second)}}
	delay := service.decisionDelay()
	if delay < 2*time.Second || delay > 4*time.Second {
		t.Fatalf("unexpected decision delay %s", delay)
	}
}

// TestNeedInteractionsPrioritizeEnergy verifies bounded autonomous need selection.
func TestNeedInteractionsPrioritizeEnergy(t *testing.T) {
	primary, secondary, found := needInteractions(petrecord.Pet{Level: 1, Energy: 20, Happiness: 10})
	if !found || primary != "pet_food" || secondary != "pet_drink" {
		t.Fatalf("unexpected energy need %q %q found=%v", primary, secondary, found)
	}
	primary, secondary, found = needInteractions(petrecord.Pet{Level: 1, Energy: 100, Happiness: 10})
	if !found || primary != "pet_toy" || secondary != "nest" {
		t.Fatalf("unexpected happiness need %q %q found=%v", primary, secondary, found)
	}
	if _, _, found = needInteractions(petrecord.Pet{Level: 1, Energy: 100, Happiness: 100}); found {
		t.Fatal("expected no autonomous need")
	}
}

// TestPetTickNoDueAllocatesNothing verifies the idle room-cycle hot path.
func TestPetTickNoDueAllocatesNothing(t *testing.T) {
	service, active := idleCycleFixture(t)
	now := time.Unix(100, 0)
	allocations := testing.AllocsPerRun(1000, func() {
		if err := service.Cycle(context.Background(), active, now); err != nil {
			t.Fatal(err)
		}
	})
	if allocations != 0 {
		t.Fatalf("expected zero allocations, got %.2f", allocations)
	}
}

// TestStationaryPetRejectsEveryLocomotionPath verifies plants remain anchored.
func TestStationaryPetRejectsEveryLocomotionPath(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	t.Cleanup(func() { _, _, _ = service.rooms.Close(context.Background(), active.ID()) })
	controller, _ := service.Active(active.ID(), 50)
	controller.mutex.Lock()
	controller.stationary = true
	controller.record.Energy = 0
	controller.nextDue = time.Time{}
	controller.followingPlayerID = 1
	controller.mutex.Unlock()
	service.config = petpolicy.Config{}.Normalize()
	service.source = fixedSource{value: 1}
	service.needs = &commandNeedConsumer{calls: make(chan int64, 1)}
	item := worldfurniture.Item{ID: 99, Definition: worldfurniture.Definition{InteractionType: "pet_food", Width: 1, Length: 1, AllowWalk: true}, Point: grid.MustPoint(2, 0)}
	if _, err := active.ReloadFurniture(item.ID, &item); err != nil {
		t.Fatal(err)
	}
	before, _ := active.UnitMotion(EntityKey(50))
	if err := service.Cycle(context.Background(), active, time.Unix(100, 0)); err != nil {
		t.Fatal(err)
	}
	after, _ := active.UnitMotion(EntityKey(50))
	if after.Moving || after.Position != before.Position {
		t.Fatalf("stationary pet moved from %+v to %+v", before, after)
	}
	if err := service.MovePet(active, 50, grid.MustPoint(2, 0)); !errors.Is(err, petrecord.ErrInvalidState) {
		t.Fatalf("expected directed movement rejection, got %v", err)
	}
	for _, action := range []CommandAction{{ID: 3, Mode: ActionHere}, {ID: 7, Mode: ActionFollow}, {ID: 43, Mode: ActionNeed, Need: CommandNeedFood}} {
		if err := service.ExecuteAction(context.Background(), active.ID(), 50, 1, action, petrecord.Command{RequiredLevel: 1}); !errors.Is(err, petrecord.ErrInvalidState) {
			t.Fatalf("action %+v returned %v", action, err)
		}
	}
}

// TestStationaryPetCancelsStaleContextualMovement verifies runtime-only reservations cannot bypass anchoring.
func TestStationaryPetCancelsStaleContextualMovement(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	t.Cleanup(func() { _, _, _ = service.rooms.Close(context.Background(), active.ID()) })
	controller, _ := service.Active(active.ID(), 50)
	controller.mutex.Lock()
	controller.stationary = true
	controller.commandNeed = commandNeedState{itemID: 99, actionID: 43, kind: CommandNeedFood}
	controller.mutex.Unlock()
	if err := service.Cycle(context.Background(), active, time.Unix(100, 0)); err != nil {
		t.Fatal(err)
	}
	controller.mutex.Lock()
	pending := controller.commandNeed.itemID
	controller.mutex.Unlock()
	unit, _ := active.UnitMotion(EntityKey(50))
	if pending != 0 || unit.Moving {
		t.Fatalf("pending=%d unit=%+v", pending, unit)
	}
}

// TestStationaryPetDecisionAllocatesNothing verifies the anchored due-cycle hot path.
func TestStationaryPetDecisionAllocatesNothing(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	controller, _ := service.Active(active.ID(), 50)
	service.config = petpolicy.Config{}.Normalize()
	service.source = fixedSource{value: 1}
	now := time.Unix(100, 0)
	configureStationaryPlant(service, controller, now)
	allocations := testing.AllocsPerRun(1000, func() {
		if err := service.Cycle(context.Background(), active, now); err != nil {
			t.Fatal(err)
		}
	})
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
	if allocations != 0 {
		t.Fatalf("expected zero allocations, got %.2f", allocations)
	}
}

// TestStationaryPlantProjectsDeathTransition verifies deadline changes reach the renderer.
func TestStationaryPlantProjectsDeathTransition(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	controller, _ := service.Active(active.ID(), 50)
	now := time.Unix(10_000, 0)
	createdAt, growAt, dieAt := now.Add(-3*time.Hour), now.Add(-2*time.Hour), now.Add(-time.Hour)
	controller.mutex.Lock()
	controller.record.TypeID, controller.record.State = 16, petrecord.StateRoom
	controller.record.CreatedAt, controller.record.GrowAt, controller.record.DieAt = createdAt, &growAt, &dieAt
	controller.stationary, controller.plantStage, controller.nextDue = true, 7, time.Time{}
	controller.mutex.Unlock()
	references := &petreference.Snapshot{}
	references.SpeciesPresent[16] = true
	references.Species[16] = petrecord.Species{TypeID: 16, Plant: true, Enabled: true}
	service.references = roomReferences{snapshot: references}
	service.source = fixedSource{value: 1}
	if err := service.Cycle(context.Background(), active, now); err != nil {
		t.Fatal(err)
	}
	unit, found := active.Unit(EntityKey(50))
	_, _, _ = service.rooms.Close(context.Background(), active.ID())
	if !found || !unitHasStatus(unit, "rip") || unitHasStatus(unit, "std") {
		t.Fatalf("expected dead plant appearance, unit=%+v found=%v", unit, found)
	}
}

// BenchmarkPetLookup measures direct room-local pet lookup.
func BenchmarkPetLookup(b *testing.B) {
	service, _ := idleCycleFixture(b)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, _ = service.Active(9, 50)
	}
}

// BenchmarkPetTickNoDue measures one shared room cycle with no due pet.
func BenchmarkPetTickNoDue(b *testing.B) {
	service, active := idleCycleFixture(b)
	now := time.Unix(100, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_ = service.Cycle(context.Background(), active, now)
	}
}

// BenchmarkPetDecisionIdle measures repeated due autonomous decisions on the shared world.
func BenchmarkPetDecisionIdle(b *testing.B) {
	service, active, _ := unitLookupFixture(b)
	service.config = petpolicy.Config{}.Normalize()
	service.source = fixedSource{value: 1}
	pet, _ := service.Active(active.ID(), 50)
	pet.stay = true
	now := time.Unix(100, 0)
	b.Cleanup(func() { _, _, _ = service.rooms.Close(context.Background(), active.ID()) })
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_ = service.Cycle(context.Background(), active, now.Add(time.Duration(index)*time.Minute))
	}
}

// BenchmarkPetDecisionStationary measures one due anchored plant decision.
func BenchmarkPetDecisionStationary(b *testing.B) {
	service, active, _ := unitLookupFixture(b)
	service.config = petpolicy.Config{}.Normalize()
	service.source = fixedSource{value: 1}
	pet, _ := service.Active(active.ID(), 50)
	now := time.Unix(100, 0)
	configureStationaryPlant(service, pet, now)
	b.Cleanup(func() { _, _, _ = service.rooms.Close(context.Background(), active.ID()) })
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_ = service.Cycle(context.Background(), active, now.Add(time.Duration(index)*time.Minute))
	}
}

// configureStationaryPlant installs one stable midpoint lifecycle fixture.
func configureStationaryPlant(service *Service, pet *activePet, now time.Time) {
	createdAt, growAt, dieAt := now.Add(-time.Hour), now.Add(time.Hour), now.Add(8*time.Hour)
	pet.stationary, pet.plantStage = true, 4
	pet.record.TypeID, pet.record.State, pet.record.CreatedAt = 16, petrecord.StateRoom, createdAt
	pet.record.GrowAt, pet.record.DieAt = &growAt, &dieAt
	references := &petreference.Snapshot{}
	references.SpeciesPresent[16] = true
	references.Species[16] = petrecord.Species{TypeID: 16, Plant: true, Enabled: true}
	service.references = roomReferences{snapshot: references}
}

// idleCycleFixture creates one already-loaded room generation with one non-due pet.
func idleCycleFixture(t testing.TB) (*Service, *roomlive.Room) {
	t.Helper()
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 25, AllowPets: true, AllowPetsEat: true})
	if err != nil {
		t.Fatal(err)
	}
	pet := &activePet{record: petrecord.Pet{ID: 50, OwnerPlayerID: 1}, nextDue: time.Unix(200, 0)}
	state := &roomState{pets: map[int64]*activePet{50: pet}}
	state.rebuildSnapshot()
	service := &Service{config: petpolicy.Config{}.Normalize(), source: fixedSource{}, active: map[int64]*roomState{9: state}}
	return service, active
}
