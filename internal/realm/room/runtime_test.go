package room

import (
	"context"
	"testing"
	"time"

	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentry "github.com/niflaot/pixels/internal/realm/room/entry"
	roomoccupancy "github.com/niflaot/pixels/internal/realm/room/events/occupancychanged"
	"github.com/niflaot/pixels/internal/realm/room/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
)

// TestMovementPublisherCompletesDoorExitAfterMovement verifies tick exits use standard teardown.
func TestMovementPublisherCompletesDoorExitAfterMovement(t *testing.T) {
	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create session peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter player room: %v", err)
	}
	players := playerlive.NewRegistry()
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	runtime := live.NewRegistry(nil)
	active, err := runtime.Activate(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, live.Occupant{PlayerID: 7, ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join runtime: %v", err)
	}

	publisher := newMovementPublisher(nil, players, bus.New(), nil, func() *live.Registry { return runtime })
	if err := publisher(context.Background(), active, []live.Movement{{PlayerID: 7, Exited: true}}); err != nil {
		t.Fatalf("publish exit: %v", err)
	}
	if active.Occupancy().Count != 0 {
		t.Fatalf("expected empty room, got %#v", active.Occupancy())
	}
	if _, found := player.CurrentRoom(); found {
		t.Fatal("expected player room cleared")
	}
}

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

	registry := NewLiveRegistry(local, netconn.NewRegistry(), nil, roomentry.Config{}, nil, nil)
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
	registry := NewLiveRegistry(local, netconn.NewRegistry(), nil, roomentry.Config{}, nil, nil)
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
	registry := NewLiveRegistry(local, netconn.NewRegistry(), nil, roomentry.Config{}, nil, nil)
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
