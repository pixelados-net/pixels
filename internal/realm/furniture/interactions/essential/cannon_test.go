package essential

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestCannonLocksConcurrentShots verifies activation state and cooldown guard.
func TestCannonLocksConcurrentShots(t *testing.T) {
	item := essentialItem("cannon", 2)
	active := essentialRoom(t, item, 1)
	service := &Service{}
	request := Request{PlayerID: 1, Room: active, Item: item}
	if err := service.useCannon(context.Background(), request); err != nil {
		t.Fatalf("fire cannon: %v", err)
	}
	fired, _ := active.FurnitureItem(item.ID)
	if fired.ExtraData != "1" {
		t.Fatalf("expected fired state, got %q", fired.ExtraData)
	}
	request.Item = fired
	if err := service.useCannon(context.Background(), request); err != nil {
		t.Fatalf("repeat cannon: %v", err)
	}
	repeated, _ := active.FurnitureItem(item.ID)
	if repeated.ExtraData != "1" {
		t.Fatalf("cooldown allowed second toggle: %q", repeated.ExtraData)
	}
}

// TestCannonReleasesActivatorAfterFire verifies delayed firing and cooldown cleanup.
func TestCannonReleasesActivatorAfterFire(t *testing.T) {
	item := essentialItem("cannon", 2)
	active := essentialRoom(t, item, 1)
	service := &Service{}
	request := Request{PlayerID: 1, Room: active, Item: item}
	if err := service.useCannon(context.Background(), request); err != nil {
		t.Fatalf("start cannon: %v", err)
	}
	active.RunScheduled(time.Now().Add(3 * time.Second))
	unit, _ := active.Unit(1)
	if unit.Moving {
		t.Fatal("expected stationary activator after release")
	}
	if _, err := service.cannonNotice(); err != nil {
		t.Fatalf("encode cannon notice: %v", err)
	}
	if !active.TryLockInteraction(item.ID, time.Now().Add(time.Second)) {
		t.Fatal("expected cooldown release")
	}
}

// TestCannonRoutesDistantActivator verifies a click walks to a reachable perimeter tile.
func TestCannonRoutesDistantActivator(t *testing.T) {
	item := essentialItem("cannon", 2)
	item.Point = grid.MustPoint(4, 0)
	active := essentialRoom(t, item, 1)
	service := &Service{}
	if err := service.useCannon(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); err != nil {
		t.Fatalf("use distant cannon: %v", err)
	}
	for range 8 {
		active.Tick()
		active.RunScheduled(time.Now().Add(time.Second))
	}
	updated, _ := active.FurnitureItem(item.ID)
	if updated.ExtraData != "1" {
		t.Fatalf("expected routed cannon activation, got %q", updated.ExtraData)
	}
}

// TestCannonWaitsForAutomaticApproach verifies clicks survive Nitro movement latency.
func TestCannonWaitsForAutomaticApproach(t *testing.T) {
	item := essentialItem("cannon", 2)
	item.Point = grid.MustPoint(4, 0)
	active := essentialRoom(t, item, 1)
	if _, err := active.MoveTo(1, grid.MustPoint(3, 0)); err != nil {
		t.Fatalf("start approach: %v", err)
	}
	service := &Service{}
	request := Request{PlayerID: 1, Room: active, Item: item}
	if err := service.useCannon(context.Background(), request); err != nil {
		t.Fatalf("queue cannon use: %v", err)
	}
	for range 6 {
		active.Tick()
		active.RunScheduled(time.Now().Add(time.Second))
	}
	fired, _ := active.FurnitureItem(item.ID)
	if fired.ExtraData != "1" {
		t.Fatalf("expected cannon activation after approach, got %q", fired.ExtraData)
	}
}

// TestCannonLineFollowsRotation verifies three-tile shot geometry.
func TestCannonLineFollowsRotation(t *testing.T) {
	line := cannonLine(grid.MustPoint(4, 4), worldunit.RotationEast)
	expected := []grid.Point{grid.MustPoint(4, 3), grid.MustPoint(4, 2), grid.MustPoint(4, 1)}
	if len(line) != len(expected) {
		t.Fatalf("unexpected line %#v", line)
	}
	for index := range expected {
		if line[index] != expected[index] {
			t.Fatalf("unexpected line %#v", line)
		}
	}
}
