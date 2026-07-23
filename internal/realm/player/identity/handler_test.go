package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inchange "github.com/niflaot/pixels/networking/inbound/user/name/change"
	incheck "github.com/niflaot/pixels/networking/inbound/user/name/check"
	outchange "github.com/niflaot/pixels/networking/outbound/user/name/change"
	outcheck "github.com/niflaot/pixels/networking/outbound/user/name/check"
	"github.com/niflaot/pixels/pkg/bus"
)

// failingRenameStore returns one configured rename failure.
type failingRenameStore struct {
	err error
}

// Rename returns the configured failure.
func (store failingRenameStore) Rename(context.Context, int64, string) (RenameResult, error) {
	return RenameResult{}, store.err
}

// TestIdentityHandlersCheckAndRename verifies successful packet adaptation and live projection.
func TestIdentityHandlersCheckAndRename(t *testing.T) {
	store := &renameStore{}
	service := New(store, identityFinder{available: true, taken: map[string]bool{}}, nil)
	handler, connection, packets, player := identityFixture(t, service)
	checkPacket, err := codec.NewPacket(incheck.Header, incheck.Definition, codec.String("Valid"))
	if err != nil {
		t.Fatal(err)
	}
	if err = handler.check(connection, checkPacket); err != nil {
		t.Fatalf("check: %v", err)
	}
	if len(*packets) != 1 || (*packets)[0].Header != outcheck.Header {
		t.Fatalf("check packets=%#v", *packets)
	}
	changePacket, err := codec.NewPacket(inchange.Header, inchange.Definition, codec.String("Renamed"))
	if err != nil {
		t.Fatal(err)
	}
	if err = handler.change(connection, changePacket); err != nil {
		t.Fatalf("change: %v", err)
	}
	if len(*packets) != 3 || (*packets)[1].Header != outchange.Header || player.Username() != "Renamed" || store.candidate != "Renamed" {
		t.Fatalf("packets=%#v username=%q candidate=%q", *packets, player.Username(), store.candidate)
	}
}

// TestIdentityHandlersMapFailuresAndMalformedPackets verifies stable client failures.
func TestIdentityHandlersMapFailuresAndMalformedPackets(t *testing.T) {
	handler, connection, packets, _ := identityFixture(t, New(failingRenameStore{err: ErrRenameDisabled}, identityFinder{available: true, taken: map[string]bool{}}, nil))
	packet, err := codec.NewPacket(inchange.Header, inchange.Definition, codec.String("Renamed"))
	if err != nil {
		t.Fatal(err)
	}
	if err = handler.change(connection, packet); err != nil || len(*packets) != 1 || (*packets)[0].Header != outchange.Header {
		t.Fatalf("failure packets=%#v err=%v", *packets, err)
	}
	if err = handler.check(connection, codec.Packet{Header: inchange.Header}); err == nil {
		t.Fatal("expected malformed check rejection")
	}
	if err = handler.change(connection, codec.Packet{Header: incheck.Header}); err == nil {
		t.Fatal("expected malformed change rejection")
	}
}

// TestIdentityRegistrationAndErrorCodes verifies registration and stable error mapping.
func TestIdentityRegistrationAndErrorCodes(t *testing.T) {
	RegisterHandlers(nil, Handler{})
	registry := netconn.NewHandlerRegistry()
	RegisterHandlers(registry, Handler{})
	if registry.Len() != 2 {
		t.Fatalf("handlers=%d", registry.Len())
	}
	for err, expected := range map[error]int32{ErrRenameDisabled: ResultDisabled, ErrUsernameTaken: ResultTaken, ErrReservationMissing: ResultTaken, errors.New("other"): ResultInvalid} {
		if actual := resultCode(err); actual != expected {
			t.Fatalf("error=%v code=%d", err, actual)
		}
	}
	if _, _, found := (Handler{}).player(netconn.Context{}); found {
		t.Fatal("expected missing player")
	}
	checkPacket, _ := codec.NewPacket(incheck.Header, incheck.Definition, codec.String("Valid"))
	changePacket, _ := codec.NewPacket(inchange.Header, inchange.Definition, codec.String("Valid"))
	if err := (Handler{}).check(netconn.Context{}, checkPacket); err != nil {
		t.Fatalf("unbound check: %v", err)
	}
	if err := (Handler{}).change(netconn.Context{}, changePacket); err != nil {
		t.Fatalf("unbound change: %v", err)
	}
	if err := (Handler{}).projectRoom(1, "name"); err != nil {
		t.Fatalf("empty room projection: %v", err)
	}
}

// identityFixture creates one authenticated packet-capturing identity handler.
func identityFixture(t *testing.T, service *Service) (Handler, netconn.Context, *[]codec.Packet, *playerlive.Player) {
	t.Helper()
	inbound, outbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	_ = inbound.Register(1, func(ctx netconn.Context, _ codec.Packet) error { connection = ctx; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	packets := make([]codec.Packet, 0, 2)
	session, _ := netconn.NewSession(netconn.SessionConfig{ID: "identity", Kind: "websocket", Inbound: inbound, Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	_ = session.Receive(context.Background(), codec.Packet{Header: 1})
	connections := netconn.NewRegistry()
	_ = connections.Register(session)
	peer, _ := playerlive.NewSessionPeer("identity", "websocket", time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	players := playerlive.NewRegistry()
	_ = players.Add(player)
	bindings := binding.NewRegistry()
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "identity", ConnectionKind: "websocket"})
	rooms := roomlive.NewRegistry(nil)
	active, _ := rooms.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	roomGrid, _ := grid.Parse("00", grid.WithDoor(0, 0))
	_ = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth})
	_, _ = rooms.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "identity", ConnectionKind: "websocket"})
	return Handler{Service: service, Bindings: bindings, Players: players, Rooms: rooms, Connections: connections, Events: bus.New()}, connection, &packets, player
}
