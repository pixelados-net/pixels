package walk

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestHandleMoveErrorSettlesActiveMovement verifies soft misses do not broadcast a snapping status.
func TestHandleMoveErrorSettlesActiveMovement(t *testing.T) {
	handler, player := handlerForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	connections := netconn.NewRegistry()
	sent := registeredConnectionForWalkTest(t, connections, "conn")
	handler.Connections = connections
	room, _ := handler.Runtime.Find(9)
	if _, err := room.MoveTo(7, grid.MustPoint(1, 0)); err != nil {
		t.Fatalf("start movement: %v", err)
	}
	if err := handler.handleMoveError(context.Background(), room, 7, grid.MustPoint(1, 0), worldpath.ErrInvalidGoal); err != nil {
		t.Fatalf("handle movement miss: %v", err)
	}
	if len(*sent) != 0 {
		t.Fatalf("expected no immediate snapping packet, got %#v", *sent)
	}
	movements := room.Tick()
	if len(movements) != 1 || !movements[0].Settled || movements[0].Moved {
		t.Fatalf("expected deferred neutral settlement, got %#v", movements)
	}
}
