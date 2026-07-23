package warning

import (
	"context"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// registeredSessionForTest creates one live player and transport session.
func registeredSessionForTest(t *testing.T) (*playerlive.Registry, *netconn.Registry, *int) {
	t.Helper()
	peer, err := playerlive.NewSessionPeer("test", "websocket", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "test"}, peer)
	if err != nil {
		t.Fatal(err)
	}
	players := playerlive.NewRegistry()
	if err = players.Add(player); err != nil {
		t.Fatal(err)
	}
	connections := netconn.NewRegistry()
	sent := 0
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "test", Kind: "websocket", Outbound: outbound, Sender: func(context.Context, codec.Packet) error { sent++; return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }, StartedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatal(err)
	}
	return players, connections, &sent
}

// alertStoreForTest embeds unused store behavior and records queued warnings.
type alertStoreForTest struct {
	// Store supplies unused persistence methods.
	sanctionrecord.Store
	// alert stores the queued warning.
	alert sanctionrecord.Alert
}

// QueueAlert records one offline warning.
func (store *alertStoreForTest) QueueAlert(_ context.Context, alert sanctionrecord.Alert) error {
	store.alert = alert
	return nil
}

// TestWarnQueuesOfflineDelivery verifies disconnected recipients never lose warnings.
func TestWarnQueuesOfflineDelivery(t *testing.T) {
	store := &alertStoreForTest{}
	effect := NewWarn(store, playerlive.NewRegistry(), netconn.NewRegistry())
	punishment := sanctionrecord.Punishment{ID: 9, ReceiverPlayerID: 7, Reason: "warning"}
	if err := effect.Apply(context.Background(), punishment); err != nil {
		t.Fatal(err)
	}
	if store.alert.PlayerID != 7 || store.alert.PunishmentID == nil || *store.alert.PunishmentID != 9 || store.alert.Message != "warning" {
		t.Fatalf("alert=%+v", store.alert)
	}
	if err := effect.Revoke(context.Background(), punishment); err != nil {
		t.Fatal(err)
	}
}

// TestWarnSendsOnlineWithoutQueueing verifies immediate delivery wins over persistence.
func TestWarnSendsOnlineWithoutQueueing(t *testing.T) {
	players, connections, sent := registeredSessionForTest(t)
	store := &alertStoreForTest{}
	effect := NewWarn(store, players, connections)
	if effect.Kind() != sanctionrecord.KindWarn {
		t.Fatalf("kind=%q", effect.Kind())
	}
	if err := effect.Apply(context.Background(), sanctionrecord.Punishment{ID: 9, ReceiverPlayerID: 7, Reason: "warning"}); err != nil {
		t.Fatal(err)
	}
	if *sent != 1 || store.alert.PlayerID != 0 {
		t.Fatalf("sent=%d alert=%+v", *sent, store.alert)
	}
}
