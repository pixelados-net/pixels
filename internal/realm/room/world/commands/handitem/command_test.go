package handitem

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestHandleDropsOnlyTheActorsItem verifies voluntary hand-item clearing.
func TestHandleDropsOnlyTheActorsItem(t *testing.T) {
	handler, active, actor := handFixture(t, 1)
	if _, err := active.SetHandItem(7, 21); err != nil {
		t.Fatal(err)
	}
	if _, err := active.SetHandItem(8, 35); err != nil {
		t.Fatal(err)
	}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: actor, Kind: KindDrop}}); err != nil {
		t.Fatalf("drop hand item: %v", err)
	}
	first, _ := active.Unit(7)
	second, _ := active.Unit(8)
	if first.HandItem != 0 || second.HandItem != 35 {
		t.Fatalf("unexpected hand items actor=%d target=%d", first.HandItem, second.HandItem)
	}
}

// TestHandleGivesOnlyToAnAdjacentPlayer verifies transfer and proximity guards.
func TestHandleGivesOnlyToAnAdjacentPlayer(t *testing.T) {
	handler, active, actor := handFixture(t, 2)
	if _, err := active.SetHandItem(7, 21); err != nil {
		t.Fatal(err)
	}
	target, _ := active.Unit(8)
	commandValue := Command{Handler: actor, Kind: KindGive, TargetUnitID: target.UnitID}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: commandValue}); err != nil {
		t.Fatalf("reject distant transfer: %v", err)
	}
	first, _ := active.Unit(7)
	if first.HandItem != 21 {
		t.Fatalf("distant transfer cleared actor item: %d", first.HandItem)
	}
	if _, err := active.MoveTo(8, grid.MustPoint(1, 0)); err != nil {
		t.Fatalf("move target adjacent: %v", err)
	}
	active.Tick()
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: commandValue}); err != nil {
		t.Fatalf("give hand item: %v", err)
	}
	first, _ = active.Unit(7)
	second, _ := active.Unit(8)
	if first.HandItem != 0 || second.HandItem != 21 {
		t.Fatalf("unexpected transfer actor=%d target=%d", first.HandItem, second.HandItem)
	}
}

// handFixture creates two authenticated players in one loaded room.
func handFixture(t *testing.T, targetX int) (Handler, *roomlive.Room, netconn.Context) {
	t.Helper()
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatal(err)
	}
	for _, identity := range []struct {
		id   int64
		name string
	}{
		{id: 7, name: "actor"}, {id: 8, name: "target"},
	} {
		connectionID := netconn.ID(identity.name)
		peer, peerErr := playerlive.NewSessionPeer(connectionID, "websocket", time.Now())
		if peerErr != nil {
			t.Fatal(peerErr)
		}
		player, playerErr := playerlive.NewPlayer(playerlive.Snapshot{ID: identity.id, Username: identity.name}, peer)
		if playerErr != nil {
			t.Fatal(playerErr)
		}
		if playerErr = player.EnterRoom(9); playerErr != nil {
			t.Fatal(playerErr)
		}
		if playerErr = players.Add(player); playerErr != nil {
			t.Fatal(playerErr)
		}
		if playerErr = bindings.Add(binding.Binding{PlayerID: identity.id, ConnectionID: connectionID, ConnectionKind: "websocket"}); playerErr != nil {
			t.Fatal(playerErr)
		}
		if _, playerErr = runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: identity.id, Username: identity.name, ConnectionID: connectionID, ConnectionKind: "websocket"}); playerErr != nil {
			t.Fatal(playerErr)
		}
	}
	if targetX > 0 {
		if _, err = active.MoveTo(8, grid.MustPoint(targetX, 0)); err != nil {
			t.Fatal(err)
		}
		for index := 0; index < targetX; index++ {
			active.Tick()
		}
	}
	return Handler{Players: players, Bindings: bindings, Runtime: runtime, Connections: netconn.NewRegistry()}, active, netconn.Context{ConnectionID: "actor", ConnectionKind: "websocket"}
}
