package teleport

import (
	"context"
	"sync"
	"testing"

	teleportfailed "github.com/niflaot/pixels/internal/realm/furniture/events/teleportfailed"
	teleportstarted "github.com/niflaot/pixels/internal/realm/furniture/events/teleportstarted"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestBlockedApproachReleasesTransit verifies a soft rejection remains retriable.
func TestBlockedApproachReleasesTransit(t *testing.T) {
	service, active, _ := serviceForTest(t, "teleport")
	if _, err := active.TeleportUnit(7, grid.MustPoint(0, 2), worldunit.RotationEast, false); err != nil {
		t.Fatalf("reposition player: %v", err)
	}
	if _, err := service.runtime.Join(context.Background(), active.ID(), occupantForTeleportTest(8, "two")); err != nil {
		t.Fatalf("join blocking player: %v", err)
	}
	failed, started := 0, 0
	if _, err := service.events.(*bus.Bus).Subscribe(teleportfailed.Name, bus.PriorityNormal, func(context.Context, bus.Event) error {
		failed++
		return nil
	}); err != nil {
		t.Fatalf("subscribe failed event: %v", err)
	}
	if _, err := service.events.(*bus.Bus).Subscribe(teleportstarted.Name, bus.PriorityNormal, func(context.Context, bus.Event) error {
		started++
		return nil
	}); err != nil {
		t.Fatalf("subscribe started event: %v", err)
	}
	request := StartRequest{PlayerID: 7, Room: active, ItemID: 1}
	if err := service.Start(context.Background(), request); err != nil {
		t.Fatalf("reject blocked approach: %v", err)
	}
	if failed != 1 || started != 0 {
		t.Fatalf("unexpected events failed=%d started=%d", failed, started)
	}
	if _, found := service.rooms.Load(active.ID()); found || service.reservations != nil {
		t.Fatalf("expected released attempt rooms=%v reservations=%#v", found, service.reservations)
	}
	if _, _, err := service.runtime.Leave(context.Background(), 8); err != nil {
		t.Fatalf("remove blocking player: %v", err)
	}
	if err := service.Start(context.Background(), request); err != nil {
		t.Fatalf("retry teleport: %v", err)
	}
	if started != 1 {
		t.Fatalf("expected retry to start, got %d events", started)
	}
	service.removeTransit(active.ID(), 7)
}

// TestConcurrentPairUseAllowsOneTransit verifies atomic endpoint reservation.
func TestConcurrentPairUseAllowsOneTransit(t *testing.T) {
	service, active, _ := serviceForTest(t, "teleport")
	if _, err := service.runtime.Join(context.Background(), active.ID(), occupantForTeleportTest(8, "two")); err != nil {
		t.Fatalf("join second player: %v", err)
	}
	requests := [...]StartRequest{
		{PlayerID: 7, Room: active, ItemID: 1},
		{PlayerID: 8, Room: active, ItemID: 2},
	}
	var group sync.WaitGroup
	errors := make([]error, len(requests))
	for index := range requests {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			errors[index] = service.Start(context.Background(), requests[index])
		}(index)
	}
	group.Wait()
	for index, err := range errors {
		if err != nil {
			t.Fatalf("concurrent start %d: %v", index, err)
		}
	}
	state := service.roomState(active.ID())
	state.mutex.Lock()
	count := len(state.transits)
	var winner int64
	for playerID := range state.transits {
		winner = playerID
	}
	state.mutex.Unlock()
	if count != 1 {
		t.Fatalf("expected one active pair transit, got %d", count)
	}
	service.reservationMutex.Lock()
	reserved := len(service.reservations) == 2 && service.reservations[1] == winner && service.reservations[2] == winner
	service.reservationMutex.Unlock()
	if !reserved {
		t.Fatalf("unexpected pair reservations winner=%d", winner)
	}
	service.removeTransit(active.ID(), winner)
	if service.reservations != nil {
		t.Fatalf("expected reservation release, got %#v", service.reservations)
	}
}

// TestCancelPlayerClosesAndReleasesPair verifies disconnect cleanup.
func TestCancelPlayerClosesAndReleasesPair(t *testing.T) {
	service, active, now := serviceForTest(t, "teleport")
	if err := service.Start(context.Background(), StartRequest{PlayerID: 7, Room: active, ItemID: 1}); err != nil {
		t.Fatalf("start teleport: %v", err)
	}
	if err := service.Cycle(context.Background(), active, now); err != nil {
		t.Fatalf("open teleport: %v", err)
	}
	service.cancelPlayer(context.Background(), 7)
	if item, _ := active.FurnitureItem(1); item.ExtraData != "0" {
		t.Fatalf("expected closed source after cancel, got %q", item.ExtraData)
	}
	if _, found := service.rooms.Load(active.ID()); found || service.reservations != nil {
		t.Fatalf("expected released transition rooms=%v reservations=%#v", found, service.reservations)
	}
}
