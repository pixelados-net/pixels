package live

import (
	"errors"
	"testing"
	"time"

	navviewer "github.com/niflaot/pixels/internal/realm/navigator/viewer/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestSnapshotFromRecordMapsPersistentData verifies persistent records become runtime snapshots.
func TestSnapshotFromRecordMapsPersistentData(t *testing.T) {
	homeRoomID := int64(22)
	record := playerservice.Record{
		Player: playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 7}}, Username: "ian"},
		Profile: playermodel.Profile{
			PlayerID:        7,
			Look:            "hd-180-1",
			Gender:          playermodel.GenderMale,
			Motto:           "hello",
			HomeRoomID:      &homeRoomID,
			AllowNameChange: true,
		},
	}

	snapshot := SnapshotFromRecord(record)
	if snapshot.ID != 7 || snapshot.Username != "ian" || snapshot.HomeRoomID == nil {
		t.Fatalf("unexpected snapshot: %#v", snapshot)
	}
	if !snapshot.Valid() {
		t.Fatal("expected snapshot to be valid")
	}
}

// TestNewSessionPeerValidatesInput verifies session peer creation.
func TestNewSessionPeerValidatesInput(t *testing.T) {
	now := time.Now()
	peer, err := NewSessionPeer("ws-1", "websocket", now)
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	if peer.ConnectionID() != "ws-1" || peer.ConnectionKind() != "websocket" || !peer.AuthenticatedAt().Equal(now) {
		t.Fatalf("unexpected peer: %#v", peer)
	}

	_, err = NewSessionPeer("", "websocket", now)
	if !errors.Is(err, ErrInvalidPeer) {
		t.Fatalf("expected invalid peer error, got %v", err)
	}
}

// TestNewPlayerValidatesAndExposesState verifies live player creation.
func TestNewPlayerValidatesAndExposesState(t *testing.T) {
	player := mustPlayer(t, 10, "ian")

	if player.ID() != 10 || player.Username() != "ian" {
		t.Fatalf("unexpected player: %#v", player.Snapshot())
	}
	if player.Peer().ConnectionID() != "ws-10" {
		t.Fatalf("unexpected peer: %#v", player.Peer())
	}
}

// TestPlayerReplaceSnapshotPreservesIdentity verifies snapshot replacement.
func TestPlayerReplaceSnapshotPreservesIdentity(t *testing.T) {
	player := mustPlayer(t, 10, "ian")

	err := player.ReplaceSnapshot(Snapshot{ID: 10, Username: "ianfedev", Motto: "updated"})
	if err != nil {
		t.Fatalf("replace snapshot: %v", err)
	}
	if player.Username() != "ianfedev" {
		t.Fatalf("expected updated username, got %s", player.Username())
	}

	err = player.ReplaceSnapshot(Snapshot{ID: 11, Username: "other"})
	if !errors.Is(err, ErrInvalidPlayer) {
		t.Fatalf("expected invalid player error, got %v", err)
	}
}

// TestPlayerNavigatorLifecycle verifies embedded navigator viewer behavior.
func TestPlayerNavigatorLifecycle(t *testing.T) {
	player := mustPlayer(t, 10, "ian")
	if _, found := player.Navigator(); found {
		t.Fatal("expected no navigator viewer")
	}

	viewer := player.OpenNavigator()
	viewer.SetLastSearch(navSearchForTest())

	found, ok := player.Navigator()
	if !ok || found.LastSearch().Code != "hotel_view" {
		t.Fatalf("unexpected navigator viewer found=%v viewer=%#v", ok, found)
	}

	closed, ok := player.CloseNavigator()
	if !ok || closed != viewer {
		t.Fatalf("unexpected closed viewer found=%v viewer=%#v", ok, closed)
	}
	if _, found := player.Navigator(); found {
		t.Fatal("expected closed navigator viewer")
	}
}

// TestPlayerRoomPresenceLifecycle verifies embedded room presence behavior.
func TestPlayerRoomPresenceLifecycle(t *testing.T) {
	player := mustPlayer(t, 10, "ian")
	if _, found := player.CurrentRoom(); found {
		t.Fatal("expected no current room")
	}

	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	roomID, found := player.CurrentRoom()
	if !found || roomID != 9 {
		t.Fatalf("unexpected current room found=%v id=%d", found, roomID)
	}

	leftID, found := player.LeaveRoom()
	if !found || leftID != 9 {
		t.Fatalf("unexpected left room found=%v id=%d", found, leftID)
	}
}

// TestRegistryNavigatorAudienceFiltersPlayers verifies navigator audience derivation.
func TestRegistryNavigatorAudienceFiltersPlayers(t *testing.T) {
	registry := NewRegistry()
	first := mustPlayer(t, 10, "ian")
	second := mustPlayer(t, 11, "ada")
	first.OpenNavigator()
	second.OpenNavigator().SetCategoryCounts(false)
	if err := registry.Add(first); err != nil {
		t.Fatalf("add first player: %v", err)
	}
	if err := registry.Add(second); err != nil {
		t.Fatalf("add second player: %v", err)
	}

	audience := registry.NavigatorAudience()
	if len(audience) != 1 || audience[0].ID() != 10 {
		t.Fatalf("unexpected audience %#v", audience)
	}
}

// TestRegistryLifecycle verifies live player registry behavior.
func TestRegistryLifecycle(t *testing.T) {
	registry := NewRegistry()
	player := mustPlayer(t, 10, "ian")

	if err := registry.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	if err := registry.Add(player); !errors.Is(err, ErrPlayerExists) {
		t.Fatalf("expected duplicate player error, got %v", err)
	}
	if registry.Count() != 1 {
		t.Fatalf("expected count 1, got %d", registry.Count())
	}

	found, ok := registry.Find(10)
	if !ok || found.ID() != 10 {
		t.Fatalf("expected player 10, got %#v", found)
	}

	snapshot := registry.Snapshot()
	removed, ok := registry.Remove(10)
	if !ok || removed.ID() != 10 {
		t.Fatalf("expected removed player, got %#v", removed)
	}
	if len(snapshot) != 1 || registry.Count() != 0 {
		t.Fatalf("unexpected registry state snapshot=%d count=%d", len(snapshot), registry.Count())
	}
}

// mustPlayer creates a live test player.
func mustPlayer(t *testing.T, id int64, username string) *Player {
	t.Helper()

	peer, err := NewSessionPeer("ws-10", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}

	player, err := NewPlayer(Snapshot{ID: id, Username: username}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	return player
}

// navSearchForTest returns a navigator search state for tests.
func navSearchForTest() navviewer.LastSearch {
	return navviewer.LastSearch{Code: "hotel_view", Query: "demo"}
}
