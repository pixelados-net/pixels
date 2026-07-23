package session

import (
	"errors"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestPlayerResolvesBoundConnection verifies successful player resolution.
func TestPlayerResolvesBoundConnection(t *testing.T) {
	handler := netconn.Context{ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: handler.ConnectionID, ConnectionKind: handler.ConnectionKind}); err != nil {
		t.Fatalf("add binding: %v", err)
	}
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer(handler.ConnectionID, handler.ConnectionKind, time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	resolved, err := Player(handler, bindings, players)
	if err != nil {
		t.Fatalf("resolve player: %v", err)
	}
	if resolved.ID() != 7 {
		t.Fatalf("unexpected resolved player %#v", resolved)
	}
}

// TestPlayerRejectsMissingRegistries verifies the nil-registry guard.
func TestPlayerRejectsMissingRegistries(t *testing.T) {
	_, err := Player(netconn.Context{}, nil, nil)
	if !errors.Is(err, ErrBindingNotFound) {
		t.Fatalf("expected binding not found, got %v", err)
	}
}

// TestPlayerRejectsMissingBinding verifies unbound connections are rejected.
func TestPlayerRejectsMissingBinding(t *testing.T) {
	_, err := Player(netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}, binding.NewRegistry(), playerlive.NewRegistry())
	if !errors.Is(err, ErrBindingNotFound) {
		t.Fatalf("expected binding not found, got %v", err)
	}
}

// TestPlayerRejectsMissingLivePlayer verifies bound but offline players are rejected.
func TestPlayerRejectsMissingLivePlayer(t *testing.T) {
	handler := netconn.Context{ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: handler.ConnectionID, ConnectionKind: handler.ConnectionKind}); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	_, err := Player(handler, bindings, playerlive.NewRegistry())
	if !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf("expected player not found, got %v", err)
	}
}
