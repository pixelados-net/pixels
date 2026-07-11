package live

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestRoomLoopPublishesMovements verifies owner loop movement publishing.
func TestRoomLoopPublishesMovements(t *testing.T) {
	room := worldRoomForTest(t, "00", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, err := room.MoveTo(7, pointForTest(t, 1, 0)); err != nil {
		t.Fatalf("move unit: %v", err)
	}

	var calls atomic.Int32
	room.startLoop(context.Background(), time.Millisecond, func(context.Context, *Room, []Movement) error {
		calls.Add(1)
		return nil
	}, nil, 0)
	defer room.stopLoop()

	deadline := time.After(200 * time.Millisecond)
	for calls.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("expected movement publish")
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

// TestTickBroadcastsStopAfterFurnitureInvalidatesPath verifies clients receive a neutral status.
func TestTickBroadcastsStopAfterFurnitureInvalidatesPath(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	goal := pointForTest(t, 2, 0)
	chair := worldfurniture.Item{
		ID: 4,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, AllowSit: true,
			Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusSit}},
		},
		Point: goal, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(chair.ID, &chair); err != nil {
		t.Fatalf("place chair: %v", err)
	}
	if _, err := room.MoveTo(7, goal); err != nil {
		t.Fatalf("start movement: %v", err)
	}
	if _, err := room.ReloadFurniture(chair.ID, nil); err != nil {
		t.Fatalf("pick up chair: %v", err)
	}

	movements := room.Tick()
	if len(movements) != 1 || movements[0].Moved || !movements[0].Settled || movements[0].Unit.Moving {
		t.Fatalf("expected one neutral stop movement, got %#v", movements)
	}
	if hasStatus(movements[0].Unit.Statuses, worldunit.StatusMove) {
		t.Fatalf("expected move status cleared, got %#v", movements[0].Unit.Statuses)
	}
}

// TestRoomLoopIgnoresMissingPublisher verifies nil publishers do not start.
func TestRoomLoopIgnoresMissingPublisher(t *testing.T) {
	room := worldRoomForTest(t, "0", 0, 0)
	room.startLoop(context.Background(), time.Millisecond, nil, nil, 0)
	room.stopLoop()
}

// BenchmarkSweepDoorbellEmpty measures the dominant no-waiter tick path.
func BenchmarkSweepDoorbellEmpty(b *testing.B) {
	room, err := NewRoom(Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25})
	if err != nil {
		b.Fatalf("create room: %v", err)
	}
	now := time.Now()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = room.SweepDoorbell(now, 5*time.Minute)
	}
}
