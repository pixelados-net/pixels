package flatcats

import (
	"context"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inflatcats "github.com/niflaot/pixels/networking/inbound/navigator/browse/flatcats"
	outflatcats "github.com/niflaot/pixels/networking/outbound/navigator/browse/flatcats"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
)

// TestHandlerSendsFlatCategories verifies flat category request handling.
func TestHandlerSendsFlatCategories(t *testing.T) {
	sent := make([]codec.Packet, 0, 1)
	session := testSession(t, NewPacketHandler(Handler{
		Players:    testPlayers(t),
		Bindings:   testBindings(t),
		Categories: testCategories{},
	}, zap.NewNop()), &sent)

	if err := session.Receive(context.Background(), codec.Packet{Header: inflatcats.Header}); err != nil {
		t.Fatalf("receive flat cats: %v", err)
	}
	if len(sent) != 1 || sent[0].Header != outflatcats.Header {
		t.Fatalf("unexpected sent packets %#v", sent)
	}
}

// testCategories provides room categories for handler tests.
type testCategories struct{}

// ListCategories lists room categories for handler tests.
func (testCategories) ListCategories(context.Context) ([]roommodel.Category, error) {
	return []roommodel.Category{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 7}}, Caption: "Social", Visible: true}}, nil
}

// testSession creates an authenticated handler test session.
func testSession(t *testing.T, handler netconn.Handler, sent *[]codec.Packet) *netconn.Session {
	t.Helper()

	inbound := netconn.NewHandlerRegistry()
	RegisterPacketHandler(inbound, handler)
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(_ netconn.Context, _ codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       "conn",
		Kind:     "websocket",
		Inbound:  inbound,
		Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			*sent = append(*sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	connectSession(t, session)

	return session
}

// testPlayers creates a bound live player registry.
func testPlayers(t *testing.T) *playerlive.Registry {
	t.Helper()

	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("new peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 1, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("new player: %v", err)
	}
	players := playerlive.NewRegistry()
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	return players
}

// testBindings creates a player connection binding registry.
func testBindings(t *testing.T) *binding.Registry {
	t.Helper()

	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 1, ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	return bindings
}

// connectSession moves a session to connected.
func connectSession(t *testing.T, session *netconn.Session) {
	t.Helper()

	for _, event := range []netconn.Event{netconn.EventPacketReceived, netconn.EventAuthenticationStarted} {
		if err := session.Transition(event); err != nil {
			t.Fatalf("transition %s: %v", event, err)
		}
	}
	if err := session.Authenticate(time.Now()); err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if err := session.Transition(netconn.EventSessionReady); err != nil {
		t.Fatalf("session ready: %v", err)
	}
}
