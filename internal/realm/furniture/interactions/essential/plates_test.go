package essential

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestPressurePlateDerivesOccupancy verifies debounced walk-on and walk-off state.
func TestPressurePlateDerivesOccupancy(t *testing.T) {
	item := essentialItem("pressureplate", 2)
	item.Definition.AllowWalk, item.Definition.StackHeight = true, 0
	active := essentialRoom(t, item, 1)
	service := &Service{}
	if _, err := active.TeleportUnit(1, item.Point, worldunit.RotationEast, false); err != nil {
		t.Fatalf("position unit: %v", err)
	}
	service.schedulePressure(context.Background(), active, item)
	active.RunScheduled(time.Now().Add(time.Second))
	pressed, _ := active.FurnitureItem(item.ID)
	if pressed.ExtraData != "1" {
		t.Fatalf("expected pressed state, got %q", pressed.ExtraData)
	}
	if _, err := active.TeleportUnit(1, grid.MustPoint(0, 0), worldunit.RotationEast, false); err != nil {
		t.Fatalf("move unit off: %v", err)
	}
	service.schedulePressure(context.Background(), active, pressed)
	active.RunScheduled(time.Now().Add(time.Second))
	released, _ := active.FurnitureItem(item.ID)
	if released.ExtraData != "0" {
		t.Fatalf("expected released state, got %q", released.ExtraData)
	}
}

// TestColorPlateClampsDeltas verifies bounded occupancy state changes.
func TestColorPlateClampsDeltas(t *testing.T) {
	item := essentialItem("colorplate", 3)
	active := essentialRoom(t, item, 1)
	service := &Service{}
	for range 4 {
		current, _ := active.FurnitureItem(item.ID)
		if err := service.changeColorPlate(context.Background(), active, current, 1); err != nil {
			t.Fatalf("increment plate: %v", err)
		}
	}
	current, _ := active.FurnitureItem(item.ID)
	if current.ExtraData != "2" {
		t.Fatalf("expected clamped high state, got %q", current.ExtraData)
	}
	for range 4 {
		current, _ = active.FurnitureItem(item.ID)
		_ = service.changeColorPlate(context.Background(), active, current, -1)
	}
	current, _ = active.FurnitureItem(item.ID)
	if current.ExtraData != "0" {
		t.Fatalf("expected clamped low state, got %q", current.ExtraData)
	}
}

// BenchmarkColorPlate measures a movement-driven plate state update.
func BenchmarkColorPlate(b *testing.B) {
	item := essentialItem("colorplate", 5)
	active := essentialRoom(b, item, 1)
	service := &Service{}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		current, _ := active.FurnitureItem(item.ID)
		_ = service.changeColorPlate(ctx, active, current, 1)
		_, _ = active.SetFurnitureExtraData(item.ID, "0")
	}
}
