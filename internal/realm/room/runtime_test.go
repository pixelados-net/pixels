package room

import (
	"context"
	"testing"
	"time"

	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomoccupancy "github.com/niflaot/pixels/internal/realm/room/runtime/events/occupancychanged"
	"github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
)

// noopMovementPublisher accepts benchmark events without retaining them.
type noopMovementPublisher struct{}

// Publish accepts one benchmark event.
func (noopMovementPublisher) Publish(context.Context, bus.Event) error { return nil }

// TestPublishFurnitureStepsIgnoresSameItemFootprint verifies multi-tile effects do not restart per step.
func TestPublishFurnitureStepsIgnoresSameItemFootprint(t *testing.T) {
	active, err := live.NewRoom(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	item := worldfurniture.Item{ID: 12, Point: grid.MustPoint(0, 0), Definition: worldfurniture.Definition{
		InteractionType: "effect_tile", Width: 2, Length: 1, AllowWalk: true,
	}}
	if err = active.LoadWorld(live.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{item}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	events := bus.New()
	count := 0
	for _, name := range []bus.Name{"furniture.walked_on", "furniture.walked_off"} {
		if _, err = events.Subscribe(name, bus.PriorityNormal, func(context.Context, bus.Event) error { count++; return nil }); err != nil {
			t.Fatal(err)
		}
	}
	movement := live.Movement{PlayerID: 7, Unit: live.UnitSnapshot{
		Previous: worldpath.Position{Point: grid.MustPoint(0, 0)}, Position: worldpath.Position{Point: grid.MustPoint(1, 0)}, BodyRotation: worldunit.RotationSouth,
	}}
	if err = publishFurnitureSteps(context.Background(), events, active, movement); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected no same-item transition events, got %d", count)
	}
}

// BenchmarkPublishFurnitureStepsSameItem measures the multi-tile no-transition movement path.
func BenchmarkPublishFurnitureStepsSameItem(b *testing.B) {
	active, err := live.NewRoom(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		b.Fatal(err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		b.Fatal(err)
	}
	item := worldfurniture.Item{ID: 12, Point: grid.MustPoint(0, 0), Definition: worldfurniture.Definition{
		InteractionType: "effect_tile", Width: 2, Length: 1, AllowWalk: true,
	}}
	if err = active.LoadWorld(live.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{item}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		b.Fatal(err)
	}
	movement := live.Movement{PlayerID: 7, Unit: live.UnitSnapshot{
		Previous: worldpath.Position{Point: grid.MustPoint(0, 0)}, Position: worldpath.Position{Point: grid.MustPoint(1, 0)},
	}}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err = publishFurnitureSteps(ctx, noopMovementPublisher{}, active, movement); err != nil {
			b.Fatal(err)
		}
	}
}

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
