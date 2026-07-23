package admin

import (
	"context"
	"errors"
	"testing"

	petcatalog "github.com/niflaot/pixels/internal/realm/pet/catalog"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// adminReferences supplies one immutable administrative fixture.
type adminReferences struct {
	// snapshot stores the fixture generation.
	snapshot *petreference.Snapshot
	// refreshes counts explicit refreshes.
	refreshes int
}

// Current returns the fixture generation.
func (references *adminReferences) Current(context.Context) (*petreference.Snapshot, error) {
	return references.snapshot, nil
}

// Refresh records one explicit refresh.
func (references *adminReferences) Refresh(context.Context) error {
	references.refreshes++
	return nil
}

// UpdateAdmin applies one fixture appearance mutation.
func (store *adminStore) UpdateAdmin(_ context.Context, petID int64, patch petrecord.AdminPatch) (petrecord.Pet, bool, error) {
	if store.pet.ID != petID || store.pet.Version != patch.Version {
		return store.pet, false, nil
	}
	if patch.Name != nil {
		store.pet.Name = *patch.Name
	}
	if patch.Color != nil {
		store.pet.Color = *patch.Color
	}
	store.pet.Version++
	return store.pet, true, nil
}

// Grant creates or returns the fixture pet idempotently.
func (store *adminStore) Grant(_ context.Context, params petrecord.GrantParams) (petrecord.Pet, bool, error) {
	if store.pet.ID != 0 {
		return store.pet, false, nil
	}
	store.pet = petrecord.Pet{ID: 90, OwnerPlayerID: params.OwnerPlayerID, Name: params.Name, TypeID: params.TypeID, BreedID: params.BreedID, PaletteID: params.PaletteID, Color: params.Color, State: petrecord.StateInventory, Version: 1}
	return store.pet, true, nil
}

// AppendGlobalAudit records one reference refresh audit.
func (store *adminStore) AppendGlobalAudit(_ context.Context, _ int64, action string, _ string) error {
	if store.auditErr != nil {
		return store.auditErr
	}
	store.auditAction = action
	return nil
}

// TestReadReferenceViewsAreStable verifies sorted protected manifest projections.
func TestReadReferenceViewsAreStable(t *testing.T) {
	service, store, references := readFixture()
	values, err := service.List(context.Background(), petrecord.AdminFilter{Limit: 1})
	if err != nil || len(values) != 1 || values[0].ID != store.pet.ID {
		t.Fatalf("list=%+v err=%v", values, err)
	}
	if found, ok, findErr := service.Find(context.Background(), store.pet.ID); findErr != nil || !ok || found.ID != store.pet.ID {
		t.Fatalf("find=%+v ok=%v err=%v", found, ok, findErr)
	}
	species, err := service.Species(context.Background())
	if err != nil || len(species) != 2 || species[0].TypeID != 0 || species[1].TypeID != 1 {
		t.Fatalf("species=%+v err=%v", species, err)
	}
	typeID := int32(0)
	breeds, err := service.Breeds(context.Background(), &typeID, true)
	if err != nil || len(breeds) != 2 || breeds[0].BreedID != 1 || breeds[1].BreedID != 2 {
		t.Fatalf("breeds=%+v err=%v", breeds, err)
	}
	commands, links, err := service.Commands(context.Background())
	if err != nil || len(commands) != 1 || commands[0].ID != 1 || len(links[0]) != 1 {
		t.Fatalf("commands=%+v links=%+v err=%v", commands, links, err)
	}
	if err = service.RefreshReference(context.Background(), Audit{ActorPlayerID: 1, Reason: "QA"}); err != nil {
		t.Fatal(err)
	}
	if references.refreshes != 1 || store.auditAction != "reference_refreshed" {
		t.Fatalf("refreshes=%d audit=%q", references.refreshes, store.auditAction)
	}
}

// TestCreateAndUpdateValidateThenAudit verifies protected identity workflows.
func TestCreateAndUpdateValidateThenAudit(t *testing.T) {
	service, store, _ := readFixture()
	store.pet = petrecord.Pet{}
	created, fresh, err := service.Create(context.Background(), CreateParams{OwnerPlayerID: 7, Name: "Pixel", TypeID: 0, BreedID: 1, PaletteID: 1, Color: "#aabbcc", OperationKey: "request-1", Audit: Audit{ActorPlayerID: 1, Reason: "QA"}})
	if err != nil || !fresh || created.ID != 90 || created.Color != "AABBCC" || store.auditAction != "created" {
		t.Fatalf("created=%+v fresh=%v action=%q err=%v", created, fresh, store.auditAction, err)
	}
	name, color := "Renamed", "112233"
	updated, err := service.Update(context.Background(), created.ID, petrecord.AdminPatch{Name: &name, Color: &color, Version: 1}, Audit{ActorPlayerID: 1, Reason: "rename"})
	if err != nil || updated.Name != name || updated.Color != color || updated.Version != 2 || store.auditAction != "updated" {
		t.Fatalf("updated=%+v action=%q err=%v", updated, store.auditAction, err)
	}
	bad := "invalid name!"
	if _, err = service.Update(context.Background(), created.ID, petrecord.AdminPatch{Name: &bad, Version: 2}, Audit{ActorPlayerID: 1, Reason: "bad"}); !errors.Is(err, petrecord.ErrInvalidName) {
		t.Fatalf("expected invalid name, got %v", err)
	}
}

// readFixture creates one administrative service with complete catalog references.
func readFixture() (*Service, *adminStore, *adminReferences) {
	store := &adminStore{pet: petrecord.Pet{ID: 50, OwnerPlayerID: 7, Name: "Pixel", State: petrecord.StateInventory, Version: 1}}
	snapshot := &petreference.Snapshot{Breeds: map[petreference.BreedKey]petrecord.Breed{
		{TypeID: 0, BreedID: 2, PaletteID: 2}: {TypeID: 0, BreedID: 2, PaletteID: 2, Enabled: true, Sellable: true},
		{TypeID: 0, BreedID: 1, PaletteID: 1}: {TypeID: 0, BreedID: 1, PaletteID: 1, Enabled: true, Sellable: true},
	}}
	snapshot.SpeciesPresent[0], snapshot.SpeciesPresent[1] = true, true
	snapshot.Species[0], snapshot.Species[1] = petrecord.Species{TypeID: 0, Enabled: true}, petrecord.Species{TypeID: 1, Enabled: true}
	snapshot.CommandPresent[1] = true
	snapshot.Commands[1] = petrecord.Command{ID: 1, NameKey: "sit", Enabled: true}
	snapshot.SpeciesCommands[0] = []int32{1}
	references := &adminReferences{snapshot: snapshot}
	config := petpolicy.Config{Enabled: true}
	rooms := roomlive.NewRegistry(nil)
	runtimeService := petruntime.New(config, store, references, rooms, nil, nil, nil, nil, nil, nil, nil)
	catalog := petcatalog.New(config, store, references, nil, nil, runtimeService, nil, rooms, nil)
	return New(store, catalog, nil, references, runtimeService, rooms), store, references
}
