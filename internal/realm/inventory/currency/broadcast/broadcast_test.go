package broadcast

import (
	"context"
	"testing"
	"time"

	currencychanged "github.com/niflaot/pixels/internal/realm/inventory/currency/events/changed"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcredits "github.com/niflaot/pixels/networking/outbound/user/currency/credits"
	outnotification "github.com/niflaot/pixels/networking/outbound/user/currency/notification"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestHandleSendsCreditsAndSeasonalPackets verifies player-only currency projection.
func TestHandleSendsCreditsAndSeasonalPackets(t *testing.T) {
	broadcaster, connection := broadcasterForTest(t)

	err := broadcaster.Handle(context.Background(), bus.Event{Payload: currencychanged.Payload{
		PlayerID: 7, CurrencyType: -1, Amount: 100, Delta: 10,
	}})
	if err != nil {
		t.Fatalf("broadcast credits: %v", err)
	}
	err = broadcaster.Handle(context.Background(), bus.Event{Payload: currencychanged.Payload{
		PlayerID: 7, CurrencyType: 5, Amount: 20, Delta: 2,
	}})
	if err != nil {
		t.Fatalf("broadcast seasonal: %v", err)
	}

	if len(connection.sent) != 2 || connection.sent[0].Header != outcredits.Header || connection.sent[1].Header != outnotification.Header {
		t.Fatalf("unexpected packets %#v", connection.sent)
	}
}

// TestHandleSkipsOfflineAndInvalidEvents verifies absent recipients are harmless.
func TestHandleSkipsOfflineAndInvalidEvents(t *testing.T) {
	broadcaster := New(playerlive.NewRegistry(), netconn.NewRegistry())
	if err := broadcaster.Handle(context.Background(), bus.Event{Payload: "invalid"}); err != nil {
		t.Fatalf("invalid event: %v", err)
	}
	if err := broadcaster.Handle(context.Background(), bus.Event{Payload: currencychanged.Payload{PlayerID: 7}}); err != nil {
		t.Fatalf("offline event: %v", err)
	}
}

// broadcasterForTest creates one live player and connection.
func broadcasterForTest(t *testing.T) (*Broadcaster, *testConnection) {
	t.Helper()
	players := playerlive.NewRegistry()
	connections := netconn.NewRegistry()
	connection := &testConnection{id: "currency-live", done: make(chan struct{})}
	if err := connections.Register(connection); err != nil {
		t.Fatalf("register connection: %v", err)
	}
	peer, err := playerlive.NewSessionPeer(connection.id, connection.Kind(), time.Now())
	if err != nil {
		t.Fatalf("new peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("new player: %v", err)
	}
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	return New(players, connections), connection
}

// testConnection records currency packets.
type testConnection struct {
	// id identifies the connection.
	id netconn.ID
	// sent stores projected packets.
	sent []codec.Packet
	// done closes on disposal.
	done chan struct{}
}

// ID returns the connection id.
func (connection *testConnection) ID() netconn.ID { return connection.id }

// Kind returns the connection kind.
func (connection *testConnection) Kind() netconn.Kind { return "websocket" }

// StartedAt returns the connection start time.
func (connection *testConnection) StartedAt() time.Time { return time.Now() }

// AuthenticatedAt returns the connection authentication time.
func (connection *testConnection) AuthenticatedAt() (time.Time, bool) { return time.Now(), true }

// Authenticate marks the connection authenticated.
func (connection *testConnection) Authenticate(time.Time) error { return nil }

// State returns the connection state.
func (connection *testConnection) State() netconn.State { return netconn.StateConnected }

// Receive accepts an inbound packet.
func (connection *testConnection) Receive(context.Context, codec.Packet) error { return nil }

// Send records one outbound packet.
func (connection *testConnection) Send(_ context.Context, packet codec.Packet) error {
	connection.sent = append(connection.sent, packet)
	return nil
}

// Disconnect closes the connection.
func (connection *testConnection) Disconnect(context.Context, netconn.Reason) error {
	close(connection.done)
	return nil
}

// Done returns the disposal signal.
func (connection *testConnection) Done() <-chan struct{} { return connection.done }
