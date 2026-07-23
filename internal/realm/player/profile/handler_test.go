package profile

import (
	"context"
	"errors"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininfo "github.com/niflaot/pixels/networking/inbound/user/info/request"
	infigure "github.com/niflaot/pixels/networking/inbound/user/profile/figure"
	inmotto "github.com/niflaot/pixels/networking/inbound/user/profile/motto"
	inrespect "github.com/niflaot/pixels/networking/inbound/user/profile/respect"
	intags "github.com/niflaot/pixels/networking/inbound/user/profile/tags"
	outinfo "github.com/niflaot/pixels/networking/outbound/user/info"
	outfigure "github.com/niflaot/pixels/networking/outbound/user/profile/figure"
	outrespect "github.com/niflaot/pixels/networking/outbound/user/profile/respect"
	outtags "github.com/niflaot/pixels/networking/outbound/user/profile/tags"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestProfileHandlersProjectCommittedBehavior verifies all public-profile packet adapters.
func TestProfileHandlersProjectCommittedBehavior(t *testing.T) {
	store := &profileStore{tags: []string{"builder"}, result: RespectResult{Applied: true, TotalReceived: 9, Remaining: 2}}
	admin := &profileAdmin{}
	handler, connection, packets, actor, active := profileFixture(t, New(store, admin))
	infoPacket, _ := codec.NewPacket(ininfo.Header, ininfo.Definition)
	if err := handler.info(connection, infoPacket); err != nil || lastHeader(*packets) != outinfo.Header {
		t.Fatalf("info packets=%#v err=%v", *packets, err)
	}
	figurePacket, _ := codec.NewPacket(infigure.Header, infigure.Definition, codec.String("M"), codec.String("hd-180-1"))
	if err := handler.figure(connection, figurePacket); err != nil || actor.Snapshot().Look != "hd-180-1" || containsHeader(*packets, outfigure.Header) == false {
		t.Fatalf("figure snapshot=%#v packets=%#v err=%v", actor.Snapshot(), *packets, err)
	}
	mottoPacket, _ := codec.NewPacket(inmotto.Header, inmotto.Definition, codec.String("Pixels"))
	if err := handler.motto(connection, mottoPacket); err != nil || actor.Snapshot().Motto != "Pixels" {
		t.Fatalf("motto snapshot=%#v err=%v", actor.Snapshot(), err)
	}
	unit, _ := active.Unit(7)
	tagsPacket, _ := codec.NewPacket(intags.Header, intags.Definition, codec.Int32(int32(unit.UnitID)))
	if err := handler.tags(connection, tagsPacket); err != nil || lastHeader(*packets) != outtags.Header {
		t.Fatalf("tags packets=%#v err=%v", *packets, err)
	}
	respectPacket, _ := codec.NewPacket(inrespect.Header, inrespect.Definition, codec.Int32(8))
	if err := handler.respect(connection, respectPacket); err != nil || lastHeader(*packets) != outrespect.Header || actor.Snapshot().RespectsRemaining != 2 {
		t.Fatalf("respect snapshot=%#v packets=%#v err=%v", actor.Snapshot(), *packets, err)
	}
}

// TestProfileHandlersRejectMalformedAndInvalidMutations verifies safe adapter failures.
func TestProfileHandlersRejectMalformedAndInvalidMutations(t *testing.T) {
	handler, connection, packets, _, _ := profileFixture(t, New(&profileStore{}, &profileAdmin{}))
	figurePacket, _ := codec.NewPacket(infigure.Header, infigure.Definition, codec.String("X"), codec.String("bad"))
	if err := handler.figure(connection, figurePacket); err != nil || len(*packets) != 1 {
		t.Fatalf("invalid figure packets=%#v err=%v", *packets, err)
	}
	longMotto := string(make([]rune, DefaultConfig().MottoMaximumRunes+1))
	mottoPacket, _ := codec.NewPacket(inmotto.Header, inmotto.Definition, codec.String(longMotto))
	if err := handler.motto(connection, mottoPacket); err != nil || len(*packets) != 2 {
		t.Fatalf("invalid motto packets=%#v err=%v", *packets, err)
	}
	for _, call := range []func() error{
		func() error { return handler.info(connection, codec.Packet{Header: infigure.Header}) },
		func() error { return handler.figure(connection, codec.Packet{Header: ininfo.Header}) },
		func() error { return handler.motto(connection, codec.Packet{Header: ininfo.Header}) },
		func() error { return handler.tags(connection, codec.Packet{Header: ininfo.Header}) },
		func() error { return handler.respect(connection, codec.Packet{Header: ininfo.Header}) },
	} {
		if err := call(); err == nil {
			t.Fatal("expected malformed packet rejection")
		}
	}
}

// TestMottoMutationFailureKeepsSessionConnected verifies storage failures become client feedback.
func TestMottoMutationFailureKeepsSessionConnected(t *testing.T) {
	handler, connection, packets, actor, _ := profileFixture(t, New(&profileStore{}, &profileAdmin{err: errors.New("storage unavailable")}))
	packet, _ := codec.NewPacket(inmotto.Header, inmotto.Definition, codec.String("Pixels"))
	if err := handler.motto(connection, packet); err != nil {
		t.Fatalf("motto handler disconnected: %v", err)
	}
	if len(*packets) != 1 || actor.Snapshot().Motto != "" {
		t.Fatalf("unexpected packets=%+v snapshot=%+v", *packets, actor.Snapshot())
	}
}

// TestProfileRespectClientFailures verifies duplicate, exhausted, self, and remote feedback.
func TestProfileRespectClientFailures(t *testing.T) {
	results := []struct {
		name   string
		result RespectResult
		target int32
	}{
		{name: "duplicate", result: RespectResult{Duplicate: true}, target: 8},
		{name: "exhausted", target: 8},
		{name: "self", target: 7},
		{name: "remote", target: 99},
	}
	for _, current := range results {
		t.Run(current.name, func(t *testing.T) {
			handler, connection, packets, _, _ := profileFixture(t, New(&profileStore{result: current.result}, &profileAdmin{}))
			packet, _ := codec.NewPacket(inrespect.Header, inrespect.Definition, codec.Int32(current.target))
			if err := handler.respect(connection, packet); err != nil || len(*packets) != 1 {
				t.Fatalf("packets=%#v err=%v", *packets, err)
			}
		})
	}
}

// TestProfileRegistrationAndEmptyProjection verifies registration and absent-runtime behavior.
func TestProfileRegistrationAndEmptyProjection(t *testing.T) {
	RegisterHandlers(nil, Handler{})
	registry := netconn.NewHandlerRegistry()
	RegisterHandlers(registry, Handler{})
	if registry.Len() != 5 {
		t.Fatalf("handlers=%d", registry.Len())
	}
	if _, _, found := (Handler{}).player(netconn.Context{}); found {
		t.Fatal("expected missing player")
	}
	if err := (Handler{}).project(1, "look", "M", "motto"); err != nil {
		t.Fatalf("project: %v", err)
	}
	if err := (Handler{}).publishUpdated(1, true, true); err != nil {
		t.Fatalf("publish: %v", err)
	}
}

// profileFixture creates an authenticated room session with two player occupants.
func profileFixture(t *testing.T, service *Service) (Handler, netconn.Context, *[]codec.Packet, *playerlive.Player, *live.Room) {
	t.Helper()
	inbound, outbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	_ = inbound.Register(1, func(ctx netconn.Context, _ codec.Packet) error { connection = ctx; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	packets := make([]codec.Packet, 0, 12)
	session, _ := netconn.NewSession(netconn.SessionConfig{ID: "profile", Kind: "websocket", Inbound: inbound, Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	_ = session.Receive(context.Background(), codec.Packet{Header: 1})
	connections := netconn.NewRegistry()
	_ = connections.Register(session)
	peer, _ := playerlive.NewSessionPeer("profile", "websocket", time.Now())
	actor, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo", Look: "hd-180-1", Gender: playermodel.GenderMale}, peer)
	players := playerlive.NewRegistry()
	_ = players.Add(actor)
	bindings := binding.NewRegistry()
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "profile", ConnectionKind: "websocket"})
	rooms := live.NewRegistry(nil)
	active, _ := rooms.Activate(live.Snapshot{ID: 9, MaxUsers: 5})
	roomGrid, _ := grid.Parse("00", grid.WithDoor(0, 0))
	_ = active.LoadWorld(live.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth})
	_, _ = rooms.Join(context.Background(), 9, live.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "profile", ConnectionKind: "websocket"})
	_, _ = rooms.Join(context.Background(), 9, live.Occupant{PlayerID: 8, Username: "target", ConnectionID: "target", ConnectionKind: "websocket"})
	return Handler{Service: service, Bindings: bindings, Players: players, Rooms: rooms, Connections: connections, Events: bus.New()}, connection, &packets, actor, active
}

// lastHeader returns the latest captured packet header.
func lastHeader(packets []codec.Packet) uint16 {
	if len(packets) == 0 {
		return 0
	}
	return packets[len(packets)-1].Header
}

// containsHeader reports whether one packet header was captured.
func containsHeader(packets []codec.Packet, header uint16) bool {
	for _, packet := range packets {
		if packet.Header == header {
			return true
		}
	}
	return false
}
