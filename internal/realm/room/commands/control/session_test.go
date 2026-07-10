package control

import (
	"context"
	"errors"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestTargetInRoomValidatesActivePresence verifies forged protocol targets are rejected.
func TestTargetInRoomValidatesActivePresence(t *testing.T) {
	runtime := roomlive.NewRegistry(nil)
	_, err := runtime.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	_, err = runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 2, Username: "Alice", ConnectionID: "conn", ConnectionKind: netconn.Kind("websocket")})
	if err != nil {
		t.Fatalf("join room: %v", err)
	}
	if err := TargetInRoom(runtime, 9, 2); err != nil {
		t.Fatalf("expected active target: %v", err)
	}
	if err := TargetInRoom(runtime, 9, 3); !errors.Is(err, ErrTargetNotInRoom) {
		t.Fatalf("expected missing target, got %v", err)
	}
}

// TestActorResolvesBoundPlayerAndCurrentRoom verifies command session projection.
func TestActorResolvesBoundPlayerAndCurrentRoom(t *testing.T) {
	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 2, Username: "Alice"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	players := playerlive.NewRegistry()
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 2, ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}
	resolved, roomID, err := Actor(netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}, bindings, players)
	if err != nil || resolved.ID() != 2 || roomID != 9 {
		t.Fatalf("unexpected actor=%v room=%d err=%v", resolved, roomID, err)
	}
	if err := MatchRoom(9, 9); err != nil {
		t.Fatalf("match room: %v", err)
	}
	if err := MatchRoom(9, 10); !errors.Is(err, ErrRoomMismatch) {
		t.Fatalf("expected room mismatch, got %v", err)
	}
	player.LeaveRoom()
	if _, _, err := Actor(netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}, bindings, players); !errors.Is(err, ErrPlayerNotInRoom) {
		t.Fatalf("expected player not in room, got %v", err)
	}
	if _, _, err := Actor(netconn.Context{}, bindings, players); err == nil {
		t.Fatal("expected missing binding error")
	}
}
