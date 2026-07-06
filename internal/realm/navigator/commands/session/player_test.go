package session

import (
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestPlayerResolvesBoundLivePlayer verifies connection to player lookup.
func TestPlayerResolvesBoundLivePlayer(t *testing.T) {
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	handler := netconn.Context{ConnectionID: "abc", ConnectionKind: "websocket"}
	player := playerForTest(t, 7, "demo", "abc", "websocket")
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "abc", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	found, sessionBinding, err := Player(handler, bindings, players)
	if err != nil {
		t.Fatalf("resolve player: %v", err)
	}
	if found.ID() != 7 || sessionBinding.PlayerID != 7 {
		t.Fatalf("unexpected player %d binding %#v", found.ID(), sessionBinding)
	}
}

// playerForTest creates a live player for tests.
func playerForTest(t *testing.T, id int64, username string, connectionID netconn.ID, kind netconn.Kind) *playerlive.Player {
	t.Helper()

	peer, err := playerlive.NewSessionPeer(connectionID, kind, time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: id, Username: username}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	return player
}
