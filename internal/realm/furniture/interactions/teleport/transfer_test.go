package teleport

import (
	"context"
	"errors"
	"testing"
	"time"

	teleportfailed "github.com/niflaot/pixels/internal/realm/furniture/events/teleportfailed"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
)

// TestCrossRoomTransitionForwardsAndConsumesSpawn verifies the Nitro navigation handoff.
func TestCrossRoomTransitionForwardsAndConsumesSpawn(t *testing.T) {
	service, source, target, sent, now := crossRoomServiceForTest(t)
	if err := service.Start(context.Background(), StartRequest{PlayerID: 7, Room: source, ItemID: 1}); err != nil {
		t.Fatalf("start cross-room teleport: %v", err)
	}
	if err := service.Cycle(context.Background(), source, now); err != nil {
		t.Fatalf("open source: %v", err)
	}
	source.Tick()
	if err := service.Cycle(context.Background(), source, now); err != nil {
		t.Fatalf("animate source entry: %v", err)
	}
	source.Tick()
	if err := service.Cycle(context.Background(), source, now.Add(phaseDelay)); err != nil {
		t.Fatalf("settle source entry: %v", err)
	}
	if err := service.Cycle(context.Background(), source, now.Add(2*phaseDelay)); err != nil {
		t.Fatalf("show source departure: %v", err)
	}
	if item, _ := source.FurnitureItem(1); item.ExtraData != "2" {
		t.Fatalf("expected visible source departure, got %q", item.ExtraData)
	}
	if err := service.Cycle(context.Background(), source, now.Add(3*phaseDelay)); err != nil {
		t.Fatalf("forward target: %v", err)
	}
	if len(*sent) != 4 || (*sent)[3].Header != outforward.Header {
		t.Fatalf("expected room forward packet, got %#v", *sent)
	}
	_, _, _ = service.runtime.Leave(context.Background(), 7)
	if _, err := service.runtime.Join(context.Background(), 10, occupantForTeleportTest(7, "one")); err != nil {
		t.Fatalf("join target room: %v", err)
	}
	if err := service.entered(context.Background(), roomentered.Payload{PlayerID: 7, RoomID: 10}); err != nil {
		t.Fatalf("consume target spawn: %v", err)
	}
	unit, found := target.Unit(7)
	if !found || unit.Position.Point != grid.MustPoint(3, 1) {
		t.Fatalf("unexpected destination unit %#v found=%v", unit, found)
	}
	if len(*sent) != 4 {
		t.Fatalf("expected destination visuals to wait for bootstrap, got %#v", *sent)
	}
	if err := service.Cycle(context.Background(), target, now.Add(phaseDelay)); err != nil {
		t.Fatalf("show destination arrival: %v", err)
	}
	if item, _ := target.FurnitureItem(2); item.ExtraData != "2" {
		t.Fatalf("expected opened destination, got %q", item.ExtraData)
	}
	if len(*sent) != 6 || (*sent)[4].Header != outupdate.Header || (*sent)[5].Header != outstatus.Header {
		t.Fatalf("expected delayed destination visuals, got %#v", *sent)
	}
	if err := service.Cycle(context.Background(), target, now.Add(2*phaseDelay)); err != nil {
		t.Fatalf("begin destination exit: %v", err)
	}
	unit, _ = target.Unit(7)
	if item, _ := target.FurnitureItem(2); item.ExtraData != "1" || !unit.Moving {
		t.Fatalf("expected open walking exit, item=%q unit=%#v", item.ExtraData, unit)
	}
	target.Tick()
	if err := service.Cycle(context.Background(), target, now.Add(3*phaseDelay)); err != nil {
		t.Fatalf("animate destination exit: %v", err)
	}
	unit, _ = target.Unit(7)
	if unit.Position.Point != grid.MustPoint(3, 2) || !unitHasStatus(unit, worldunit.StatusMove) {
		t.Fatalf("expected visible destination movement, got %#v", unit)
	}
	target.Tick()
	if err := service.Cycle(context.Background(), target, now.Add(4*phaseDelay)); err != nil {
		t.Fatalf("settle destination exit: %v", err)
	}
	unit, _ = target.Unit(7)
	if item, _ := target.FurnitureItem(2); item.ExtraData != "0" || unit.Position.Point != grid.MustPoint(3, 2) || unit.Moving {
		t.Fatalf("expected closed settled exit, item=%q unit=%#v", item.ExtraData, unit)
	}
	if _, found := service.consumePending(7, 10); found {
		t.Fatal("expected one-time destination consumption")
	}
}

// TestPendingExpiryClearAndFailure verifies bounded transfer cleanup paths.
func TestPendingExpiryClearAndFailure(t *testing.T) {
	service, active, _, _, now := crossRoomServiceForTest(t)
	service.pending = map[int64]pendingDestination{
		7: {roomID: 10, expiresAt: now.Add(-time.Second)},
		8: {roomID: 11, expiresAt: now.Add(time.Second)},
	}
	if _, found := service.consumePending(7, 10); found {
		t.Fatal("expected expired destination rejection")
	}
	if _, found := service.consumePending(8, 10); found {
		t.Fatal("expected destination room mismatch")
	}
	service.clearPending(8)
	if service.pending != nil {
		t.Fatalf("expected empty pending map, got %#v", service.pending)
	}
	failed := false
	_, err := service.events.(*bus.Bus).Subscribe(teleportfailed.Name, bus.PriorityNormal, func(_ context.Context, event bus.Event) error {
		_, failed = event.Payload.(teleportfailed.Payload)
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe failure event: %v", err)
	}
	if err := service.fail(context.Background(), active.ID(), Transit{PlayerID: 7, Source: worldfurniture.Item{ID: 1}}, "test"); err != nil || !failed {
		t.Fatalf("expected failure event failed=%v err=%v", failed, err)
	}
}

// TestRegisterCleansPendingOnDisconnect verifies lifecycle event integration.
func TestRegisterCleansPendingOnDisconnect(t *testing.T) {
	service, active, _ := serviceForTest(t, "teleport_tile")
	local := bus.New()
	service.events = local
	lifecycle := fxtest.NewLifecycle(t)
	if err := Register(lifecycle, local, service.runtime, service); err != nil {
		t.Fatalf("register teleport lifecycle: %v", err)
	}
	service.pending = map[int64]pendingDestination{7: {roomID: active.ID(), expiresAt: time.Now().Add(time.Minute)}}
	if err := local.Publish(context.Background(), bus.Event{Name: playerdisconnected.Name, Payload: playerdisconnected.Payload{PlayerID: 7}}); err != nil {
		t.Fatalf("publish disconnect: %v", err)
	}
	if service.pending != nil {
		t.Fatalf("expected disconnect cleanup, got %#v", service.pending)
	}
	if err := local.Publish(context.Background(), bus.Event{Name: furniturewalkedon.Name, Payload: furniturewalkedon.Payload{
		PlayerID: 7, ItemID: 1, RoomID: active.ID(),
	}}); err != nil {
		t.Fatalf("publish walked-on event: %v", err)
	}
	if _, found := service.rooms.Load(active.ID()); !found {
		t.Fatal("expected walk-on subscription to start tile transition")
	}
	lifecycle.RequireStop()
}

// TestStartValidationAndRemoveTransit verifies cheap invalid and reservation cleanup paths.
func TestStartValidationAndRemoveTransit(t *testing.T) {
	service, active, _ := serviceForTest(t, "teleport")
	if err := service.Start(context.Background(), StartRequest{}); err != ErrInvalidUse {
		t.Fatalf("expected invalid use, got %v", err)
	}
	if err := service.Start(context.Background(), StartRequest{PlayerID: 7, Room: active, ItemID: 99}); err != ErrNotTeleport {
		t.Fatalf("expected non-teleport rejection, got %v", err)
	}
	state := service.roomState(active.ID())
	state.transits[7] = Transit{PlayerID: 7}
	service.removeTransit(active.ID(), 7)
	if _, found := service.rooms.Load(active.ID()); found {
		t.Fatal("expected empty room shard removal")
	}
}

// TestTransferNoOpAndMissingConnectionBranches verifies inexpensive defensive paths.
func TestTransferNoOpAndMissingConnectionBranches(t *testing.T) {
	service, active, _ := serviceForTest(t, "teleport")
	if err := service.entered(context.Background(), roomentered.Payload{PlayerID: 7, RoomID: active.ID()}); err != nil {
		t.Fatalf("unexpected entry without destination error: %v", err)
	}
	if _, found := service.playerConnection(active, 99); found {
		t.Fatal("expected missing room connection")
	}
	if err := service.forward(context.Background(), active, Transit{
		PlayerID: 99, Source: worldfurniture.Item{ID: 1}, TargetRoomID: 10,
	}); err != nil {
		t.Fatalf("missing connection should publish a soft failure: %v", err)
	}
	service.events = nil
	if err := service.publishStarted(context.Background(), active.ID(), Transit{}); err != nil {
		t.Fatalf("nil publisher should be a no-op: %v", err)
	}
	expected := errors.New("first")
	if firstError(expected, errors.New("second")) != expected || firstError(nil, expected) != expected {
		t.Fatal("unexpected first-error selection")
	}
}
