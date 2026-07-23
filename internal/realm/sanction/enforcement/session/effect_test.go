package session

import (
	"context"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// livePlayerForTest creates one registry-backed authenticated player.
func livePlayerForTest(t *testing.T, id int64) (*playerlive.Registry, *playerlive.Player) {
	t.Helper()
	peer, err := playerlive.NewSessionPeer("test", "websocket", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: id, Username: "test", AllowTrade: true}, peer)
	if err != nil {
		t.Fatal(err)
	}
	registry := playerlive.NewRegistry()
	if err = registry.Add(player); err != nil {
		t.Fatal(err)
	}
	return registry, player
}

// registeredSessionForTest creates one live player and transport session.
func registeredSessionForTest(t *testing.T) (*playerlive.Registry, *netconn.Registry, *netconn.Reason, *int) {
	t.Helper()
	players, _ := livePlayerForTest(t, 7)
	connections := netconn.NewRegistry()
	reason := &netconn.Reason{}
	sent := 0
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "test", Kind: "websocket", Outbound: outbound, Sender: func(context.Context, codec.Packet) error { sent++; return nil }, Disposer: func(_ context.Context, value netconn.Reason) error { *reason = value; return nil }, StartedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatal(err)
	}
	return players, connections, reason, &sent
}

// TestSessionDisconnectEffectsUseDistinctReasons verifies online ban and kick disposal.
func TestSessionDisconnectEffectsUseDistinctReasons(t *testing.T) {
	for _, test := range []struct {
		name string
		kind sanctionrecord.Kind
		code netconn.DisconnectCode
	}{
		{name: "ban", kind: sanctionrecord.KindBan, code: netconn.DisconnectBanned},
		{name: "kick", kind: sanctionrecord.KindKick, code: netconn.DisconnectKicked},
	} {
		t.Run(test.name, func(t *testing.T) {
			players, connections, reason, _ := registeredSessionForTest(t)
			var effect *SessionDisconnect
			if test.kind == sanctionrecord.KindBan {
				effect = &NewBan(players, connections).SessionDisconnect
			} else {
				effect = &NewKick(players, connections).SessionDisconnect
			}
			if effect.Kind() != test.kind {
				t.Fatalf("kind=%q", effect.Kind())
			}
			if err := effect.Apply(context.Background(), sanctionrecord.Punishment{ReceiverPlayerID: 7, Reason: "reason"}); err != nil {
				t.Fatal(err)
			}
			if reason.Code != test.code || reason.Message != "reason" || connections.CountAll() != 0 {
				t.Fatalf("reason=%+v connections=%d", reason, connections.CountAll())
			}
		})
	}
}

// TestSessionEffectsAreNoopsForOfflinePlayers verifies idempotent offline behavior.
func TestSessionEffectsAreNoopsForOfflinePlayers(t *testing.T) {
	registry := playerlive.NewRegistry()
	connections := netconn.NewRegistry()
	for _, effect := range []*SessionDisconnect{{kind: sanctionrecord.KindBan, players: registry, connections: connections}, {kind: sanctionrecord.KindKick, players: registry, connections: connections}} {
		if err := effect.Apply(context.Background(), sanctionrecord.Punishment{ReceiverPlayerID: 7}); err != nil {
			t.Fatal(err)
		}
		if err := effect.Revoke(context.Background(), sanctionrecord.Punishment{}); err != nil {
			t.Fatal(err)
		}
	}
}
