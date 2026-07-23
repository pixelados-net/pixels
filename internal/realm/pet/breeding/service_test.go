package breeding

import (
	"context"
	"errors"
	"testing"
	"time"

	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
)

// breedingReferences supplies one immutable breeding fixture.
type breedingReferences struct {
	// snapshot stores the fixture generation.
	snapshot *petreference.Snapshot
}

// Current returns the fixture generation.
func (references breedingReferences) Current(context.Context) (*petreference.Snapshot, error) {
	return references.snapshot, nil
}

// Refresh leaves the immutable fixture unchanged.
func (references breedingReferences) Refresh(context.Context) error { return nil }

// breedingStore returns one placed parent pair.
type breedingStore struct {
	petrecord.Store
	// pets stores the room fixture.
	pets []petrecord.Pet
	// session stores one nest-owned workflow.
	session petrecord.BreedingSession
	// offspring stores the idempotent workflow reward.
	offspring petrecord.Pet
}

// Room returns the placed parent pair.
func (store breedingStore) Room(context.Context, int64) ([]petrecord.Pet, error) {
	return store.pets, nil
}

// breedingClock supplies one deterministic lifecycle instant.
type breedingClock struct {
	// now stores the fixture time.
	now time.Time
}

// Now returns the fixture time.
func (clock breedingClock) Now() time.Time { return clock.now }

// TestParentsValidatesConsentAgeAndCompatibility verifies every server-side parent gate.
func TestParentsValidatesConsentAgeAndCompatibility(t *testing.T) {
	service, runtimeService, rooms, active, references, now := breedingFixture(t)
	first, second, err := service.parents(context.Background(), 17, 7, 1, 2)
	if err != nil || first.ID != 1 || second.ID != 2 {
		t.Fatalf("parents=%+v %+v err=%v", first, second, err)
	}
	second.PublicBreed = false
	runtimeService.ReplacePlaced(second)
	if _, _, err = service.parents(context.Background(), 17, 7, 1, 2); !errors.Is(err, petrecord.ErrNoRights) {
		t.Fatalf("expected consent rejection, got %v", err)
	}
	second.PublicBreed = true
	second.CreatedAt = now.Add(-time.Hour)
	runtimeService.ReplacePlaced(second)
	if _, _, err = service.parents(context.Background(), 17, 7, 1, 2); !errors.Is(err, petrecord.ErrInvalidState) {
		t.Fatalf("expected minimum-age rejection, got %v", err)
	}
	second.CreatedAt = now.Add(-100 * time.Hour)
	runtimeService.ReplacePlaced(second)
	delete(references.snapshot.BreedingRules, petreference.BreedingKey{ParentOneTypeID: 0, ParentTwoTypeID: 1})
	if _, _, err = service.parents(context.Background(), 17, 7, 1, 2); !errors.Is(err, petrecord.ErrInvalidState) {
		t.Fatalf("expected compatibility rejection, got %v", err)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestRarityCategoriesAreStableAndBounded verifies Nitro dialog grouping is deterministic.
func TestRarityCategoriesAreStableAndBounded(t *testing.T) {
	service, _, rooms, active, _, _ := breedingFixture(t)
	categories, err := service.rarityCategories(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != 2 || categories[0].Chance != 100 || categories[1].Chance != 33 {
		t.Fatalf("categories=%+v", categories)
	}
	if len(categories[0].Breeds) != 1 || categories[0].Breeds[0] != 1 || len(categories[1].Breeds) != 1 || categories[1].Breeds[0] != 4 {
		t.Fatalf("unexpected category breeds %+v", categories)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// breedingFixture creates two compatible adult parents in one loaded room.
func breedingFixture(t testing.TB) (*Service, *petruntime.Service, *roomlive.Registry, *roomlive.Room, breedingReferences, time.Time) {
	t.Helper()
	now := time.Unix(20_000, 0)
	roomID, firstX, secondX, y, z, rotation := int64(17), 1, 2, 0, 0.0, int16(2)
	first := petrecord.Pet{ID: 1, OwnerPlayerID: 7, Name: "First", TypeID: 0, RoomID: &roomID, X: &firstX, Y: &y, Z: &z, Rotation: &rotation, State: petrecord.StateRoom, CanBreed: true, PublicBreed: true, CreatedAt: now.Add(-100 * time.Hour), StatsAt: now, Energy: 100, Happiness: 100, Version: 1}
	second := petrecord.Pet{ID: 2, OwnerPlayerID: 8, Name: "Second", TypeID: 1, RoomID: &roomID, X: &secondX, Y: &y, Z: &z, Rotation: &rotation, State: petrecord.StateRoom, CanBreed: true, PublicBreed: true, CreatedAt: now.Add(-100 * time.Hour), StatsAt: now, Energy: 100, Happiness: 100, Version: 1}
	snapshot := &petreference.Snapshot{Breeds: map[petreference.BreedKey]petrecord.Breed{
		{TypeID: 2, BreedID: 1, PaletteID: 1}: {TypeID: 2, BreedID: 1, PaletteID: 1, Rarity: 0, Enabled: true},
		{TypeID: 2, BreedID: 4, PaletteID: 1}: {TypeID: 2, BreedID: 4, PaletteID: 1, Rarity: 2, Enabled: true},
	}, BreedingRules: map[petreference.BreedingKey]petrecord.BreedingRule{
		{ParentOneTypeID: 0, ParentTwoTypeID: 1}: {ParentOneTypeID: 0, ParentTwoTypeID: 1, ResultTypeID: 2, Enabled: true},
	}}
	for _, typeID := range []int32{0, 1, 2} {
		snapshot.SpeciesPresent[typeID] = true
		snapshot.Species[typeID] = petrecord.Species{TypeID: typeID, Breedable: true, Enabled: true}
	}
	references := breedingReferences{snapshot: snapshot}
	store := &breedingStore{pets: []petrecord.Pet{first, second}}
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: roomID, OwnerPlayerID: 7, MaxUsers: 25, AllowPets: true})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	nest := worldfurniture.Item{ID: 700, OwnerPlayerID: 7, Definition: worldfurniture.Definition{Width: 1, Length: 1, InteractionType: "pet_breeding_nest", AllowWalk: true}, Point: grid.MustPoint(0, 0)}
	if _, err = active.ReloadFurniture(nest.ID, &nest); err != nil {
		t.Fatal(err)
	}
	config := petpolicy.Config{Enabled: true, BreedingMinimumAge: 72 * time.Hour, BreedingTimeout: 2 * time.Minute}
	runtimeService := petruntime.New(config, store, references, rooms, nil, nil, nil, nil, nil, nil, nil)
	runtimeService.SetClock(breedingClock{now: now})
	if err = runtimeService.EnsureRoom(context.Background(), active); err != nil {
		t.Fatal(err)
	}
	return New(config, store, references, runtimeService, rooms, nil, nil, nil), runtimeService, rooms, active, references, now
}
