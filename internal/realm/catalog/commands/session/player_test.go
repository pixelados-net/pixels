package session

import (
	"errors"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestPlayerResolvesCatalogSession verifies bound player resolution.
func TestPlayerResolvesCatalogSession(t *testing.T) {
	connection := netconn.Context{ConnectionID: "catalog", ConnectionKind: "websocket"}
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	peer, _ := playerlive.NewSessionPeer(connection.ConnectionID, connection.ConnectionKind, time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	_ = players.Add(player)
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: connection.ConnectionID, ConnectionKind: connection.ConnectionKind})
	resolved, err := Player(connection, bindings, players)
	if err != nil || resolved.ID() != 7 {
		t.Fatalf("unexpected player %#v error %v", resolved, err)
	}
}

// TestHasClubAppliesLiveExpiration verifies catalog entitlement projection.
func TestHasClubAppliesLiveExpiration(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	peer, _ := playerlive.NewSessionPeer("club", "websocket", time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{
		ID: 8, Username: "club", Club: playermodel.Club{Level: playermodel.ClubLevelHC, ExpiresAt: &expiresAt},
	}, peer)
	if !HasClub(player) {
		t.Fatal("expected active club")
	}
	expiresAt = time.Now().Add(-time.Second)
	player, _ = playerlive.NewPlayer(playerlive.Snapshot{
		ID: 8, Username: "expired", Club: playermodel.Club{Level: playermodel.ClubLevelHC, ExpiresAt: &expiresAt},
	}, peer)
	if HasClub(player) {
		t.Fatal("expected expired club rejection")
	}
}

// TestPlayerReportsMissingCatalogSessionState verifies resolution failures.
func TestPlayerReportsMissingCatalogSessionState(t *testing.T) {
	if _, err := Player(netconn.Context{}, nil, nil); !errors.Is(err, ErrBindingNotFound) {
		t.Fatalf("expected missing binding, got %v", err)
	}
	bindings := binding.NewRegistry()
	connection := netconn.Context{ConnectionID: "catalog", ConnectionKind: "websocket"}
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: connection.ConnectionID, ConnectionKind: connection.ConnectionKind})
	if _, err := Player(connection, bindings, playerlive.NewRegistry()); !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf("expected missing player, got %v", err)
	}
}
