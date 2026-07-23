package catalog

import (
	"context"
	"errors"
	"reflect"
	"testing"

	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
)

// catalogReferences supplies one immutable catalog fixture.
type catalogReferences struct {
	// snapshot stores the fixture generation.
	snapshot *petreference.Snapshot
}

// Current returns the fixture generation.
func (references catalogReferences) Current(context.Context) (*petreference.Snapshot, error) {
	return references.snapshot, nil
}

// Refresh leaves the immutable fixture unchanged.
func (references catalogReferences) Refresh(context.Context) error { return nil }

// catalogStore captures trusted pet grants.
type catalogStore struct {
	petrecord.Store
	// params stores the latest grant request.
	params petrecord.GrantParams
	// calls counts durable grant attempts.
	calls int
}

// Grant captures and returns one idempotent fixture grant.
func (store *catalogStore) Grant(_ context.Context, params petrecord.GrantParams) (petrecord.Pet, bool, error) {
	store.params = params
	store.calls++
	return petrecord.Pet{ID: 81, OwnerPlayerID: params.OwnerPlayerID, Name: params.Name, TypeID: params.TypeID, BreedID: params.BreedID, PaletteID: params.PaletteID, Color: params.Color}, true, nil
}

// TestParsePurchaseDataValidatesExactNitroShape verifies catalog anti-cheat parsing.
func TestParsePurchaseDataValidatesExactNitroShape(t *testing.T) {
	name, paletteID, color, err := parsePurchaseData("Pixel\n2\n#aabbcc")
	if err != nil || name != "Pixel" || paletteID != 2 || color != "AABBCC" {
		t.Fatalf("unexpected purchase data %q %d %q err=%v", name, paletteID, color, err)
	}
	for _, value := range []string{"Pixel\n2", "Pixel\n-1\nFFFFFF", "Pixel\nx\nFFFFFF", "Pixel\n2\nXYZ"} {
		if _, _, _, parseErr := parsePurchaseData(value); parseErr == nil {
			t.Fatalf("expected malformed data rejection for %q", value)
		}
	}
}

// TestTypeFromProductCodeValidatesSpeciesRange verifies trusted product parsing.
func TestTypeFromProductCodeValidatesSpeciesRange(t *testing.T) {
	for value, expected := range map[string]int32{"pet0": 0, "a15": 15, "pet35": 35} {
		actual, err := typeFromProductCode(value)
		if err != nil || actual != expected {
			t.Fatalf("product %q: expected %d, got %d err=%v", value, expected, actual, err)
		}
	}
	for _, value := range []string{"pet", "petx", "pet36"} {
		if _, err := typeFromProductCode(value); !errors.Is(err, petrecord.ErrInvalidAppearance) {
			t.Fatalf("expected invalid appearance for %q, got %v", value, err)
		}
	}
}

// TestSellablePaletteSelectsStableLowestBreed verifies map order cannot alter grants.
func TestSellablePaletteSelectsStableLowestBreed(t *testing.T) {
	snapshot := &petreference.Snapshot{Breeds: map[petreference.BreedKey]petrecord.Breed{
		{TypeID: 1, BreedID: 4, PaletteID: 2}: {TypeID: 1, BreedID: 4, PaletteID: 2, Enabled: true, Sellable: true},
		{TypeID: 1, BreedID: 2, PaletteID: 2}: {TypeID: 1, BreedID: 2, PaletteID: 2, Enabled: true, Sellable: true},
		{TypeID: 1, BreedID: 1, PaletteID: 3}: {TypeID: 1, BreedID: 1, PaletteID: 3, Enabled: true, Sellable: true},
	}}
	breed, found := sellablePalette(snapshot, 1, 2)
	if !found || breed.BreedID != 2 {
		t.Fatalf("unexpected selected breed %+v found=%v", breed, found)
	}
	if _, found = sellablePalette(snapshot, 1, 9); found {
		t.Fatal("expected unknown palette rejection")
	}
}

// TestGrantCatalogValidatesAndPersistsTypedReward verifies the atomic catalog boundary.
func TestGrantCatalogValidatesAndPersistsTypedReward(t *testing.T) {
	store := &catalogStore{}
	references := catalogFixtureReferences()
	service := New(petpolicy.Config{Enabled: true}, store, references, nil, nil, nil, nil, nil, nil)
	reward, err := service.GrantCatalog(context.Background(), catalogservice.PetGrantParams{OwnerPlayerID: 7, TypeID: 0, ProductCode: "pet0", ExtraData: "Pixel\n2\n#aabbcc", OperationKey: "catalog:500"})
	if err != nil {
		t.Fatal(err)
	}
	if reward.ID != 81 || reward.OwnerPlayerID != 7 || store.calls != 1 {
		t.Fatalf("reward=%+v calls=%d", reward, store.calls)
	}
	expected := petrecord.GrantParams{OwnerPlayerID: 7, Name: "Pixel", TypeID: 0, BreedID: 3, PaletteID: 2, Color: "AABBCC", OperationKey: "catalog:500"}
	if !reflect.DeepEqual(store.params, expected) {
		t.Fatalf("params=%+v expected=%+v", store.params, expected)
	}
}

// TestGrantCatalogRejectsForgedProductAndPalette verifies client data cannot select another species.
func TestGrantCatalogRejectsForgedProductAndPalette(t *testing.T) {
	store := &catalogStore{}
	service := New(petpolicy.Config{Enabled: true}, store, catalogFixtureReferences(), nil, nil, nil, nil, nil, nil)
	values := []catalogservice.PetGrantParams{
		{OwnerPlayerID: 7, TypeID: 0, ProductCode: "pet1", ExtraData: "Pixel\n2\nAABBCC", OperationKey: "forged:1"},
		{OwnerPlayerID: 7, TypeID: 0, ProductCode: "pet0", ExtraData: "Pixel\n9\nAABBCC", OperationKey: "forged:2"},
		{OwnerPlayerID: 7, TypeID: 0, ProductCode: "pet0", ExtraData: "bad name!\n2\nAABBCC", OperationKey: "forged:3"},
	}
	for _, params := range values {
		if _, err := service.GrantCatalog(context.Background(), params); err == nil {
			t.Fatalf("expected rejection for %+v", params)
		}
	}
	if store.calls != 0 {
		t.Fatalf("invalid rewards reached persistence %d times", store.calls)
	}
}

// TestGrantRejectsDisabledSpeciesAndUnmarketableBreed verifies trusted grants still validate references.
func TestGrantRejectsDisabledSpeciesAndUnmarketableBreed(t *testing.T) {
	store := &catalogStore{}
	references := catalogFixtureReferences()
	references.snapshot.Species[0].Enabled = false
	service := New(petpolicy.Config{Enabled: true}, store, references, nil, nil, nil, nil, nil, nil)
	if _, _, err := service.Grant(context.Background(), 7, 0, 3, 2, "AABBCC", "Pixel", "admin:1"); !errors.Is(err, petrecord.ErrInvalidAppearance) {
		t.Fatalf("expected disabled species rejection, got %v", err)
	}
	references.snapshot.Species[0].Enabled = true
	breed := references.snapshot.Breeds[petreference.BreedKey{TypeID: 0, BreedID: 3, PaletteID: 2}]
	breed.Sellable = false
	references.snapshot.Breeds[petreference.BreedKey{TypeID: 0, BreedID: 3, PaletteID: 2}] = breed
	if _, _, err := service.Grant(context.Background(), 7, 0, 3, 2, "AABBCC", "Pixel", "admin:2"); !errors.Is(err, petrecord.ErrInvalidAppearance) {
		t.Fatalf("expected unavailable breed rejection, got %v", err)
	}
	if store.calls != 0 {
		t.Fatalf("invalid grants reached persistence %d times", store.calls)
	}
}

// catalogFixtureReferences creates one enabled species and sellable palette.
func catalogFixtureReferences() catalogReferences {
	snapshot := &petreference.Snapshot{Breeds: map[petreference.BreedKey]petrecord.Breed{
		{TypeID: 0, BreedID: 3, PaletteID: 2}: {TypeID: 0, BreedID: 3, PaletteID: 2, Enabled: true, Sellable: true},
	}}
	snapshot.SpeciesPresent[0] = true
	snapshot.Species[0] = petrecord.Species{TypeID: 0, Enabled: true}
	return catalogReferences{snapshot: snapshot}
}
