package runtime

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// roomReferences returns one immutable species generation.
type roomReferences struct {
	// snapshot stores the configured generation.
	snapshot *petreference.Snapshot
}

// TestMonsterPlantProjection verifies Nitro receives Arcturus-compatible appearance and lifecycle state.
func TestMonsterPlantProjection(t *testing.T) {
	pet := petrecord.Pet{ID: 50, TypeID: 16, BreedID: 1, PaletteID: 1, Color: "68A83B", Posture: "std", Parts: []petrecord.AppearancePart{
		{LayerID: 0, PartID: -1, PaletteID: 10}, {LayerID: 1, PartID: 6, PaletteID: 4},
		{LayerID: 2, PartID: 8, PaletteID: 1}, {LayerID: 3, PartID: 10, PaletteID: 7}, {LayerID: 4, PartID: 12, PaletteID: 9},
	}}
	figure := petdata.FigureString(InventoryPet(pet).Figure)
	if figure != "16 0 FFFFFF 5 0 -1 10 1 6 4 2 8 1 3 10 7 4 12 9" {
		t.Fatalf("unexpected monsterplant figure %q", figure)
	}
	dog := petrecord.Pet{TypeID: 0, BreedID: 2, PaletteID: 3, Color: "D5B35B"}
	if unchanged := InventoryPet(dog).Figure; unchanged.PaletteID != dog.PaletteID || unchanged.Color != dog.Color {
		t.Fatalf("unexpected normal pet figure %+v", unchanged)
	}
	species := petrecord.Species{TypeID: 16, Plant: true}
	for _, test := range []struct {
		name  string
		state petrecord.PlantState
		want  string
	}{{name: "growing", state: petrecord.PlantState{GrowthStage: 4}, want: "grw4"}, {name: "grown", state: petrecord.PlantState{GrowthStage: 7}, want: "std"}, {name: "dead", state: petrecord.PlantState{GrowthStage: 7, Dead: true}, want: "rip"}} {
		t.Run(test.name, func(t *testing.T) {
			if posture := RoomUnit(pet, roomlive.UnitSnapshot{}, species, test.state, false).Posture; posture != test.want {
				t.Fatalf("expected posture %q, got %q", test.want, posture)
			}
		})
	}
}

// Current returns the configured generation.
func (references roomReferences) Current(context.Context) (*petreference.Snapshot, error) {
	return references.snapshot, nil
}

// Refresh leaves the configured generation unchanged.
func (roomReferences) Refresh(context.Context) error { return nil }

// roomLoadStore controls and counts durable room reads.
type roomLoadStore struct {
	petrecord.Store
	// calls counts durable room reads.
	calls atomic.Int32
	// started reports that the first read began.
	started chan struct{}
	// release allows the read to finish.
	release chan struct{}
}

// Room returns an empty room after the test releases the load.
func (store *roomLoadStore) Room(context.Context, int64) ([]petrecord.Pet, error) {
	if store.calls.Add(1) == 1 {
		close(store.started)
	}
	<-store.release
	return nil, nil
}

// TestEnsureRoomCoalescesConcurrentLoads verifies one durable read owns a room generation.
func TestEnsureRoomCoalescesConcurrentLoads(t *testing.T) {
	store := &roomLoadStore{started: make(chan struct{}), release: make(chan struct{})}
	service := &Service{store: store, source: fixedSource{}, active: make(map[int64]*roomState)}
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 71, MaxUsers: 25, AllowPets: true})
	if err != nil {
		t.Fatal(err)
	}
	errors := make(chan error, 2)
	var workers sync.WaitGroup
	workers.Add(2)
	for range 2 {
		go func() {
			defer workers.Done()
			errors <- service.EnsureRoom(context.Background(), active)
		}()
	}
	<-store.started
	close(store.release)
	workers.Wait()
	close(errors)
	for err = range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if calls := store.calls.Load(); calls != 1 {
		t.Fatalf("expected one durable room load, got %d", calls)
	}
}

// TestEnsureRoomMarksPlantStationary verifies species capabilities are cached on room load.
func TestEnsureRoomMarksPlantStationary(t *testing.T) {
	roomID := int64(71)
	x, y, z, rotation := 1, 0, 0.0, int16(2)
	now, createdAt := time.Unix(100, 0), time.Unix(0, 0)
	growAt, dieAt := time.Unix(200, 0), time.Unix(300, 0)
	pet := petrecord.Pet{ID: 50, OwnerPlayerID: 1, TypeID: 16, RoomID: &roomID, X: &x, Y: &y, Z: &z, Rotation: &rotation, State: petrecord.StateRoom, StatsAt: now, CreatedAt: createdAt, GrowAt: &growAt, DieAt: &dieAt}
	references := &petreference.Snapshot{}
	references.SpeciesPresent[16] = true
	references.Species[16] = petrecord.Species{TypeID: 16, Plant: true, Enabled: true}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: roomID, MaxUsers: 25, AllowPets: true})
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	service := &Service{config: petpolicy.Config{}.Normalize(), store: benchmarkRoomStore{pets: []petrecord.Pet{pet}}, references: roomReferences{snapshot: references}, source: fixedSource{}, clock: actionClock{now: now}, active: make(map[int64]*roomState)}
	if err = service.EnsureRoom(context.Background(), active); err != nil {
		t.Fatal(err)
	}
	controller, found := service.Active(roomID, pet.ID)
	if !found || !controller.stationary {
		t.Fatalf("controller=%+v found=%v", controller, found)
	}
	unit, found := active.Unit(EntityKey(pet.ID))
	if !found || !unitHasStatus(unit, "grw4") {
		t.Fatalf("expected midpoint plant appearance, unit=%+v found=%v", unit, found)
	}
	derived := pet.DerivePlantState(now, references.Species[16])
	if visible := RoomUnit(pet, unit, references.Species[16], derived, false).PetLevel; visible != 4 {
		t.Fatalf("expected visible growth stage 4, got %d", visible)
	}
	placed := pet
	placedX := 2
	placed.ID, placed.X = 51, &placedX
	if _, err = active.AddEntity(EntityKey(placed.ID), placed.OwnerPlayerID, worldunit.KindPet, worldpath.Position{Point: grid.MustPoint(2, 0)}, worldunit.RotationSouth); err != nil {
		t.Fatal(err)
	}
	service.AddPlaced(context.Background(), placed)
	controller, found = service.Active(roomID, placed.ID)
	if !found || !controller.stationary {
		t.Fatalf("placed controller=%+v found=%v", controller, found)
	}
}

// TestDeadPlantDoesNotVocalize verifies expired plants never enter room chat.
func TestDeadPlantDoesNotVocalize(t *testing.T) {
	service, active, _ := unitLookupFixture(t)
	now := time.Unix(1_000, 0)
	growAt, dieAt := now.Add(-2*time.Hour), now.Add(-time.Hour)
	pet := petrecord.Pet{ID: 50, TypeID: 16, State: petrecord.StateRoom, CreatedAt: now.Add(-3 * time.Hour), GrowAt: &growAt, DieAt: &dieAt}
	references := &petreference.Snapshot{}
	references.SpeciesPresent[16] = true
	references.Species[16] = petrecord.Species{TypeID: 16, Plant: true, Enabled: true}
	references.Vocals[16] = []petrecord.Vocal{{TypeID: 16, TextKey: "pet.vocal.plant", Weight: 1, Cooldown: time.Minute, Enabled: true}}
	service.references, service.clock = roomReferences{snapshot: references}, actionClock{now: now}
	if _, spoken, err := service.vocalize(context.Background(), active, pet); err != nil || spoken {
		t.Fatalf("spoken=%v err=%v", spoken, err)
	}
}

// benchmarkRoomStore returns one prebuilt placed-pet batch.
type benchmarkRoomStore struct {
	petrecord.Store
	// pets stores placed aggregates.
	pets []petrecord.Pet
}

// Room returns the prebuilt placed-pet batch.
func (store benchmarkRoomStore) Room(context.Context, int64) ([]petrecord.Pet, error) {
	return store.pets, nil
}

// BenchmarkRoomLoadPets10 measures one ten-pet room activation.
func BenchmarkRoomLoadPets10(b *testing.B) { benchmarkRoomLoad(b, 10) }

// BenchmarkRoomLoadPets50 measures one fifty-pet room activation.
func BenchmarkRoomLoadPets50(b *testing.B) { benchmarkRoomLoad(b, 50) }

// benchmarkRoomLoad measures full controller and world-unit publication.
func benchmarkRoomLoad(b *testing.B, count int) {
	roomID := int64(71)
	now := time.Unix(100, 0)
	pets := make([]petrecord.Pet, count)
	for index := range pets {
		x, y, z, rotation := index%10, index/10, 0.0, int16(2)
		pets[index] = petrecord.Pet{ID: int64(index + 1), OwnerPlayerID: 1, RoomID: &roomID, X: &x, Y: &y, Z: &z, Rotation: &rotation, State: petrecord.StateRoom, StatsAt: now, Energy: 100, Happiness: 100, Version: 1}
	}
	roomGrid, err := grid.Parse("0000000000\r0000000000\r0000000000\r0000000000\r0000000000", grid.WithDoor(0, 0))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		active, createErr := roomlive.NewRoom(roomlive.Snapshot{ID: roomID, MaxUsers: 25, AllowPets: true})
		if createErr != nil {
			b.Fatal(createErr)
		}
		if createErr = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); createErr != nil {
			b.Fatal(createErr)
		}
		service := &Service{config: petpolicy.Config{}.Normalize(), store: benchmarkRoomStore{pets: pets}, source: fixedSource{}, clock: actionClock{now: now}, active: make(map[int64]*roomState)}
		if createErr = service.EnsureRoom(context.Background(), active); createErr != nil {
			b.Fatal(createErr)
		}
	}
}
