package engine

import (
	"context"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestProjectSuppressesPartialProgressNotification verifies ordinary deltas do not become unseen achievements in Nitro.
func TestProjectSuppressesPartialProgressNotification(t *testing.T) {
	players, connections, packets := projectionPlayer(t)
	projector := &LiveProjector{players: players, connections: connections}
	projector.Project(context.Background(), Transition{
		PlayerID: 7,
		Definition: progressionrecord.AchievementDefinition{
			ID: 7, Name: "RoomEntry", Category: "explore",
			Levels: []progressionrecord.AchievementLevel{{Level: 1, ProgressNeeded: 5}},
		},
		Mutation: progressionrecord.ProgressMutation{
			After: progressionrecord.PlayerAchievement{PlayerID: 7, DefinitionID: 7, Progress: 3},
		},
	})
	if len(*packets) != 0 {
		t.Fatalf("partial progress projected packets %#v", *packets)
	}
}

// projectionPlayer creates one online player with a packet-capturing connection.
func projectionPlayer(t testing.TB) (*playerlive.Registry, *netconn.Registry, *[]codec.Packet) {
	t.Helper()
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("progression", "websocket", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatal(err)
	}
	if err = players.Add(player); err != nil {
		t.Fatal(err)
	}
	connections := netconn.NewRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	packets := make([]codec.Packet, 0, 1)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "progression", Kind: "websocket", Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatal(err)
	}
	return players, connections, &packets
}
