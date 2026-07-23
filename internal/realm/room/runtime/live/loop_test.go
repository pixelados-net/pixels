package live

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestRoomControlledInteractionStepEntersBlockedFurniture verifies owner-controlled interaction movement.
func TestRoomControlledInteractionStepEntersBlockedFurniture(t *testing.T) {
	room := worldRoomForTest(t, "00", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	item := worldfurniture.Item{ID: 5, Point: pointForTest(t, 1, 0), Definition: worldfurniture.Definition{Width: 1, Length: 1}}
	if _, err := room.ReloadFurniture(5, &item); err != nil {
		t.Fatalf("place blocking interaction: %v", err)
	}
	if err := room.StepControlledOntoInteraction(7, item.Point, worldunit.ControlTeleporting); err != nil {
		t.Fatalf("step onto interaction: %v", err)
	}
	before, _ := room.Unit(7)
	if before.Position.Point != pointForTest(t, 0, 0) || !before.Moving {
		t.Fatalf("expected pending visible step, got %#v", before)
	}
	movements := room.Tick()
	if len(movements) != 1 || movements[0].Unit.Position.Point != item.Point || !movements[0].Moved {
		t.Fatalf("expected completed interaction step, got %#v", movements)
	}
	room.Tick()
	blocker := worldfurniture.Item{ID: 6, Point: pointForTest(t, 0, 0), Definition: worldfurniture.Definition{Width: 1, Length: 1}}
	if _, err := room.ReloadFurniture(6, &blocker); err != nil {
		t.Fatalf("place blocked exit: %v", err)
	}
	if err := room.StepControlledFromInteraction(7, blocker.Point, worldunit.ControlTeleporting); err != nil {
		t.Fatalf("step out through blocked interaction edge: %v", err)
	}
	movements = room.Tick()
	if len(movements) != 1 || movements[0].Unit.Position.Point != blocker.Point || !movements[0].Moved {
		t.Fatalf("expected forced interaction exit, got %#v", movements)
	}
}

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
	}, nil, nil, 0)
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

// TestRoomStopMovementSettlesProjectedStep verifies cancellation does not snap a moving unit.
func TestRoomStopMovementSettlesProjectedStep(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, err := room.MoveTo(7, pointForTest(t, 2, 0)); err != nil {
		t.Fatalf("move unit: %v", err)
	}
	first := room.Tick()
	if len(first) != 1 || !first[0].Moved {
		t.Fatalf("expected projected first step, got %#v", first)
	}
	if stopped, err := room.StopMovement(7); err != nil || !stopped {
		t.Fatalf("stop movement: stopped=%v err=%v", stopped, err)
	}
	settled := room.Tick()
	if len(settled) != 1 || !settled[0].Settled || settled[0].Moved || settled[0].Unit.Moving {
		t.Fatalf("expected neutral settlement, got %#v", settled)
	}
	if settled[0].Unit.Position.Point != pointForTest(t, 1, 0) {
		t.Fatalf("expected projected tile preserved, got %#v", settled[0].Unit.Position)
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
	room.startLoop(context.Background(), time.Millisecond, nil, nil, nil, 0)
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
