package grid

import "testing"

// TestGridAccessorsVerifyBounds verifies grid point access.
func TestGridAccessorsVerifyBounds(t *testing.T) {
	roomGrid, err := Parse("01\r2x", WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}

	if _, ok := roomGrid.Tile(MustPoint(9, 9)); ok {
		t.Fatal("expected missing tile outside bounds")
	}
	if _, ok := roomGrid.Index(MustPoint(9, 9)); ok {
		t.Fatal("expected missing index outside bounds")
	}
	if _, ok := roomGrid.FlagsAt(MustPoint(9, 9)); ok {
		t.Fatal("expected missing flags outside bounds")
	}

	door, ok := roomGrid.Door()
	if !ok || door != MustPoint(0, 0) {
		t.Fatalf("unexpected door %#v found=%v", door, ok)
	}
}

// TestPointRejectsInvalidCoordinates verifies point coordinate validation.
func TestPointRejectsInvalidCoordinates(t *testing.T) {
	if _, ok := NewPoint(-1, 0); ok {
		t.Fatal("expected negative x to fail")
	}
	if _, ok := NewPoint(0, -1); ok {
		t.Fatal("expected negative y to fail")
	}
	if _, ok := NewPoint(1<<16, 0); ok {
		t.Fatal("expected overflow x to fail")
	}
}

// TestMustPointPanicsOnInvalidCoordinates verifies strict point creation.
func TestMustPointPanicsOnInvalidCoordinates(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()

	MustPoint(-1, 0)
}
