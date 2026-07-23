package entry

import (
	"context"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/leave"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	indesktop "github.com/niflaot/pixels/networking/inbound/session/desktop"
	"go.uber.org/zap"
)

// TestNewHandlesDesktopView verifies desktop view dispatches room leave.
func TestNewHandlesDesktopView(t *testing.T) {
	player := playerForTest(t)
	runtime := roomlive.NewRegistry(nil)
	if _, err := runtime.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 5}); err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	handler := NewDesktop(leavecmd.Handler{
		Players:  playerRegistryForTest(t, player),
		Bindings: bindingRegistryForTest(t),
		Runtime:  runtime,
	}, zap.NewNop())
	err := handler(connectionForTest(), codec.Packet{Header: indesktop.Header})
	if err != nil {
		t.Fatalf("handle desktop view: %v", err)
	}
	if room, found := runtime.Find(9); !found || room.Occupancy().Count != 0 {
		t.Fatalf("expected empty room")
	}
}

// TestNewRejectsInvalidPacket verifies decoder errors propagate.
func TestNewRejectsInvalidPacket(t *testing.T) {
	handler := NewDesktop(leavecmd.Handler{}, zap.NewNop())
	err := handler(connectionForTest(), codec.Packet{Header: indesktop.Header, Payload: []byte{1}})
	if err == nil {
		t.Fatal("expected error")
	}
}

// playerForTest creates a live player.
func playerForTest(t *testing.T) *playerlive.Player {
	t.Helper()

	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}

	return player
}

// playerRegistryForTest creates a registry with one player.
func playerRegistryForTest(t *testing.T, player *playerlive.Player) *playerlive.Registry {
	t.Helper()

	registry := playerlive.NewRegistry()
	if err := registry.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	return registry
}

// bindingRegistryForTest creates a connection binding registry.
func bindingRegistryForTest(t *testing.T) *binding.Registry {
	t.Helper()

	registry := binding.NewRegistry()
	err := registry.Add(binding.Binding{PlayerID: 7, ConnectionID: "conn", ConnectionKind: "websocket"})
	if err != nil {
		t.Fatalf("add binding: %v", err)
	}

	return registry
}

// connectionForTest creates a handler connection context.
func connectionForTest() netconn.Context {
	return netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}
}
