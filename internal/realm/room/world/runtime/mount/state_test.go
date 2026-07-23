package mount

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestStateLinksValidUnitsAndUnlinksFromEitherSide verifies the complete mount lifecycle.
func TestStateLinksValidUnitsAndUnlinksFromEitherSide(t *testing.T) {
	units := mountUnits(t)
	state := New()
	rider, mountUnit, err := state.Set(units, 7, -50, true)
	if err != nil || rider == nil || mountUnit == nil {
		t.Fatalf("rider=%v mount=%v err=%v", rider, mountUnit, err)
	}
	if linked, found := state.Linked(7); !found || linked != -50 || rider.Position() != mountUnit.Position() || rider.RenderOffset() != worldunit.RidingHeightOffset {
		t.Fatalf("linked=%d found=%v rider=%+v mount=%+v", linked, found, rider.Position(), mountUnit.Position())
	}
	state.Unlink(units, -50)
	if _, found := state.Linked(7); found || rider.RenderOffset() != 0 {
		t.Fatalf("linked=%v offset=%d", found, rider.RenderOffset())
	}
}

// TestStateRejectsConflictsAndInvalidKinds verifies one-to-one type-safe relationships.
func TestStateRejectsConflictsAndInvalidKinds(t *testing.T) {
	units := mountUnits(t)
	state := New()
	if _, _, err := state.Set(units, 7, -50, true); err != nil {
		t.Fatal(err)
	}
	if _, _, err := state.Set(units, 8, -50, true); !errors.Is(err, ErrInvalid) {
		t.Fatalf("conflict err=%v", err)
	}
	if _, _, err := state.Set(units, -50, 7, true); !errors.Is(err, ErrInvalid) {
		t.Fatalf("kind err=%v", err)
	}
	if rider, mountUnit, err := state.Set(units, 99, -50, true); err != nil || rider != nil || mountUnit != nil {
		t.Fatalf("missing rider=%v mount=%v err=%v", rider, mountUnit, err)
	}
}

// TestLinkedAllocatesNothing verifies the mounted movement lookup stays safe for room ticks.
func TestLinkedAllocatesNothing(t *testing.T) {
	state := New()
	units := mountUnits(t)
	if _, _, err := state.Set(units, 7, -50, true); err != nil {
		t.Fatal(err)
	}
	allocations := testing.AllocsPerRun(1000, func() {
		_, _ = state.Linked(7)
	})
	if allocations != 0 {
		t.Fatalf("allocations=%.2f", allocations)
	}
}

// BenchmarkMountLinkLookup measures one warmed rider movement lookup.
func BenchmarkMountLinkLookup(b *testing.B) {
	state := New()
	units := mountUnits(b)
	if _, _, err := state.Set(units, 7, -50, true); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, _ = state.Linked(7)
	}
}

// mountUnits creates two players and one pet for relationship tests.
func mountUnits(t testing.TB) map[int64]*worldunit.Unit {
	t.Helper()
	units := make(map[int64]*worldunit.Unit)
	for key, params := range map[int64]worldunit.Params{
		7:   {ID: 1, OwnerID: 7, Kind: worldunit.KindPlayer, Position: worldpath.Position{Point: grid.MustPoint(0, 0)}},
		8:   {ID: 2, OwnerID: 8, Kind: worldunit.KindPlayer, Position: worldpath.Position{Point: grid.MustPoint(0, 1)}},
		-50: {ID: 3, OwnerID: 7, Kind: worldunit.KindPet, Position: worldpath.Position{Point: grid.MustPoint(1, 0)}},
	} {
		unit, err := worldunit.New(params)
		if err != nil {
			t.Fatal(err)
		}
		units[key] = unit
	}
	return units
}
