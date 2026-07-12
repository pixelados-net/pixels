package essential

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestHandItemGivesConfiguredItem verifies carried state without durable inventory mutation.
func TestHandItemGivesConfiguredItem(t *testing.T) {
	item := essentialItem("handitem", 2)
	item.Definition.CustomParams = "27"
	active := essentialRoom(t, item, 1)
	service := &Service{random: fixedSource(0)}
	if err := service.useHandItem(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); err != nil {
		t.Fatalf("use hand item: %v", err)
	}
	unit, _ := active.Unit(1)
	if unit.HandItem != 27 {
		t.Fatalf("expected hand item 27, got %d", unit.HandItem)
	}
	active.RunScheduled(time.Now().Add(time.Second))
	updated, _ := active.FurnitureItem(item.ID)
	if updated.ExtraData != "0" {
		t.Fatalf("expected animation reset, got %q", updated.ExtraData)
	}
}

// TestVendingDeliversAfterAnimation verifies delayed random delivery.
func TestVendingDeliversAfterAnimation(t *testing.T) {
	item := essentialItem("vendingmachine", 2)
	item.Point = grid.MustPoint(2, 0)
	item.Definition.CustomParams = "2,5,7"
	active := essentialRoom(t, item, 1)
	if _, err := active.TeleportUnit(1, grid.MustPoint(3, 0), worldunit.RotationWest, false); err != nil {
		t.Fatalf("position actor: %v", err)
	}
	service := &Service{random: fixedSource(1)}
	if err := service.useHandItem(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); err != nil {
		t.Fatalf("use vending: %v", err)
	}
	active.RunScheduled(time.Now().Add(2 * time.Second))
	unit, _ := active.Unit(1)
	if unit.HandItem != 5 {
		t.Fatalf("expected deterministic vending item 5, got %d", unit.HandItem)
	}
	active.RunScheduled(time.Now().Add(3 * time.Second))
	updated, _ := active.FurnitureItem(item.ID)
	if updated.ExtraData != "0" {
		t.Fatalf("expected vending reset, got %q", updated.ExtraData)
	}
}

// TestHandItemIDsRejectsMalformedValues verifies defensive configuration parsing.
func TestHandItemIDsRejectsMalformedValues(t *testing.T) {
	items := handItemIDs("2,nope,-1,7")
	if len(items) != 2 || items[0] != 2 || items[1] != 7 {
		t.Fatalf("unexpected parsed items %#v", items)
	}
}

// TestVendingWalksToActivator verifies the shared controlled approach workflow.
func TestVendingWalksToActivator(t *testing.T) {
	item := essentialItem("vendingmachine", 2)
	item.Point = grid.MustPoint(4, 0)
	item.Rotation = worldunit.RotationWest
	item.Definition.CustomParams = "9"
	active := essentialRoom(t, item, 1)
	service := &Service{random: fixedSource(0)}
	request := Request{PlayerID: 1, Room: active, Item: item}
	if err := service.useHandItem(context.Background(), request); err != nil {
		t.Fatalf("start vending walk: %v", err)
	}
	for range 8 {
		active.Tick()
		active.RunScheduled(time.Now().Add(time.Second))
	}
	active.RunScheduled(time.Now().Add(3 * time.Second))
	unit, _ := active.Unit(1)
	if unit.HandItem != 9 {
		t.Fatalf("expected delivered item after walk, got %d", unit.HandItem)
	}
}

// TestNoSidesVendingUsesPerimeter verifies unrestricted perimeter activation.
func TestNoSidesVendingUsesPerimeter(t *testing.T) {
	item := essentialItem("vendingmachine_no_sides", 2)
	item.Point = grid.MustPoint(2, 0)
	item.Definition.CustomParams = "11"
	active := essentialRoom(t, item, 1)
	service := &Service{random: fixedSource(0)}
	if err := service.useHandItem(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); err != nil {
		t.Fatalf("use no-sides vending: %v", err)
	}
	for range 4 {
		active.Tick()
		active.RunScheduled(time.Now().Add(time.Second))
	}
	active.RunScheduled(time.Now().Add(2 * time.Second))
	unit, _ := active.Unit(1)
	if unit.HandItem != 11 {
		t.Fatalf("expected no-sides item 11, got %d", unit.HandItem)
	}
}

// TestEmptyHandItemConfigurationIsIgnored verifies malformed definitions stay harmless.
func TestEmptyHandItemConfigurationIsIgnored(t *testing.T) {
	item := essentialItem("handitem", 1)
	active := essentialRoom(t, item, 1)
	service := &Service{random: fixedSource(0)}
	if err := service.useHandItem(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); err != nil {
		t.Fatalf("use empty hand item: %v", err)
	}
	unit, _ := active.Unit(1)
	if unit.HandItem != 0 {
		t.Fatalf("unexpected hand item %d", unit.HandItem)
	}
}
