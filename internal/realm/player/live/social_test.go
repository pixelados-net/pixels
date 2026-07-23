package live

import (
	"testing"
	"time"
)

// TestPlayerProfileViewer verifies profile observation lifecycle.
func TestPlayerProfileViewer(t *testing.T) {
	peer, _ := NewSessionPeer("connection", "websocket", time.Now())
	player, _ := NewPlayer(Snapshot{ID: 7, Username: "demo"}, peer)
	if !player.ViewProfile(9) {
		t.Fatal("expected profile viewer to open")
	}
	if profileID, found := player.ViewedProfile(); !found || profileID != 9 {
		t.Fatalf("viewed profile id=%d found=%t", profileID, found)
	}
	if profileID, found := player.CloseProfile(); !found || profileID != 9 {
		t.Fatalf("closed profile id=%d found=%t", profileID, found)
	}
}

// TestPlayerIgnoredProjection verifies replacement and mutation semantics.
func TestPlayerIgnoredProjection(t *testing.T) {
	peer, _ := NewSessionPeer("connection", "websocket", time.Now())
	player, err := NewPlayer(Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("new player: %v", err)
	}
	player.ReplaceIgnored([]int64{8, 9, 7, 0})
	if !player.IsIgnoring(8) || !player.IsIgnoring(9) || player.IsIgnoring(7) {
		t.Fatal("unexpected replaced ignored projection")
	}
	if player.Ignore(8) || !player.Ignore(10) || !player.Unignore(8) || player.Unignore(8) {
		t.Fatal("unexpected ignored mutation result")
	}
}

// BenchmarkPlayerIsIgnoring measures the room-chat recipient hot path.
func BenchmarkPlayerIsIgnoring(b *testing.B) {
	peer, _ := NewSessionPeer("connection", "websocket", time.Now())
	player, _ := NewPlayer(Snapshot{ID: 7, Username: "demo"}, peer)
	player.ReplaceIgnored([]int64{8})
	b.ReportAllocs()
	for b.Loop() {
		if !player.IsIgnoring(8) {
			b.Fatal("missing ignored player")
		}
	}
}
