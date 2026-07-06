package room

import (
	"context"
	"testing"

	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	roomoccupancy "github.com/niflaot/pixels/internal/realm/room/events/occupancychanged"
	"github.com/niflaot/pixels/internal/realm/room/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
)

// TestNewLiveRegistryPublishesRoomEvents verifies occupancy event mapping.
func TestNewLiveRegistryPublishesRoomEvents(t *testing.T) {
	local := bus.New()
	var event roomoccupancy.Payload
	if _, err := local.Subscribe(roomoccupancy.Name, bus.PriorityNormal, func(_ context.Context, received bus.Event) error {
		event = received.Payload.(roomoccupancy.Payload)
		return nil
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	registry := NewLiveRegistry(local, netconn.NewRegistry())
	if _, err := registry.Activate(live.Snapshot{ID: 9, MaxUsers: 1}); err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := registry.Join(context.Background(), 9, live.Occupant{PlayerID: 7, ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	if event.RoomID != 9 || event.Count != 1 {
		t.Fatalf("unexpected event %#v", event)
	}
}

// TestRegisterRuntimeCleanupRemovesDisconnectedPlayer verifies disconnect cleanup.
func TestRegisterRuntimeCleanupRemovesDisconnectedPlayer(t *testing.T) {
	lifecycle := fxtest.NewLifecycle(t)
	local := bus.New()
	registry := NewLiveRegistry(local, netconn.NewRegistry())
	if err := RegisterRuntimeCleanup(lifecycle, local, local, registry, netconn.NewRegistry()); err != nil {
		t.Fatalf("register cleanup: %v", err)
	}
	if _, err := registry.Activate(live.Snapshot{ID: 9, MaxUsers: 1}); err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := registry.Join(context.Background(), 9, live.Occupant{PlayerID: 7, ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	err := local.Publish(context.Background(), bus.Event{Name: playerdisconnected.Name, Payload: playerdisconnected.Payload{PlayerID: 7}})
	if err != nil {
		t.Fatalf("publish disconnect: %v", err)
	}

	room, found := registry.Find(9)
	if !found || room.Occupancy().Count != 0 {
		t.Fatalf("expected empty room found=%v room=%#v", found, room)
	}
	lifecycle.RequireStop()
}

// TestRegisterRuntimeCleanupIgnoresUnknownPayload verifies cleanup type safety.
func TestRegisterRuntimeCleanupIgnoresUnknownPayload(t *testing.T) {
	lifecycle := fxtest.NewLifecycle(t)
	local := bus.New()
	registry := NewLiveRegistry(local, netconn.NewRegistry())
	if err := RegisterRuntimeCleanup(lifecycle, local, local, registry, netconn.NewRegistry()); err != nil {
		t.Fatalf("register cleanup: %v", err)
	}

	err := local.Publish(context.Background(), bus.Event{Name: playerdisconnected.Name, Payload: "ignored"})
	if err != nil {
		t.Fatalf("publish disconnect: %v", err)
	}

	if registry.Count() != 0 {
		t.Fatalf("expected empty registry, got %d", registry.Count())
	}
	lifecycle.RequireStop()
}
