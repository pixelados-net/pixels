package effect

import (
	"context"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestInventoryAndSelectionProjection verifies owner and room packet delivery.
func TestInventoryAndSelectionProjection(t *testing.T) {
	store := newMemoryStore()
	players, connections, packets := effectPlayer(t)
	rooms, active := effectRoom(t)
	service := New(store, nil, players, connections, rooms, nil)
	service.now = func() time.Time { return time.Unix(100, 0) }
	if _, err := service.Grant(context.Background(), 7, 101, 60, SourceAdmin); err != nil {
		t.Fatal(err)
	}
	if err := service.Enable(context.Background(), 7, 101); err != nil {
		t.Fatal(err)
	}
	if err := service.SendInventory(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if len(*packets) < 5 {
		t.Fatalf("expected inventory, activation, selection, and room packets, got %d", len(*packets))
	}
	player, _ := players.Find(7)
	selected := player.Snapshot().ActiveEffectID
	unit, _ := active.Unit(7)
	if selected == nil || *selected != 101 || unit.ActiveEffectID != 101 {
		t.Fatalf("selected=%v unit=%#v", selected, unit)
	}
	if err := service.Enable(context.Background(), 7, 0); err != nil {
		t.Fatal(err)
	}
	if player.Snapshot().ActiveEffectID != nil {
		t.Fatal("expected live selection to clear")
	}
}

// TestProjectionHelpersVerifyProtocolFields verifies inventory conversion branches.
func TestProjectionHelpersVerifyProtocolFields(t *testing.T) {
	now := time.Unix(100, 0)
	activated := now.Add(-10 * time.Second)
	record := listEffect(Effect{ID: 9, DurationSeconds: 60, ActivatedAt: &activated, RemainingCharges: 2}, now)
	if record.Type != 9 || record.InactiveEffectsInInventory != 1 || record.SecondsLeftIfActive != 50 || record.Permanent {
		t.Fatalf("unexpected record %#v", record)
	}
	permanent := listEffect(Effect{ID: 8, RemainingCharges: 1}, now)
	if !permanent.Permanent || permanent.InactiveEffectsInInventory != 1 {
		t.Fatalf("unexpected permanent record %#v", permanent)
	}
	service := New(newMemoryStore(), nil, nil, nil, nil, nil)
	if _, found, err := service.rankEffect(context.Background(), 7); err != nil || found {
		t.Fatalf("unexpected rank effect found=%t err=%v", found, err)
	}
}

// effectPlayer creates one online player with a packet-capturing connection.
func effectPlayer(t testing.TB) (*playerlive.Registry, *netconn.Registry, *[]codec.Packet) {
	t.Helper()
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("effect", "websocket", time.Now())
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
	packets := make([]codec.Packet, 0, 8)
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "effect", Kind: "websocket", Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatal(err)
	}
	return players, connections, &packets
}

// effectRoom creates one active room containing the effect fixture player.
func effectRoom(t testing.TB) (*roomlive.Registry, *roomlive.Room) {
	t.Helper()
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatal(err)
	}
	if _, err = rooms.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "effect", ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	return rooms, active
}
