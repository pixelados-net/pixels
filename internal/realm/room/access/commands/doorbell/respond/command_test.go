package respond

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomdoorbell "github.com/niflaot/pixels/internal/realm/room/access/doorbell"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outdoorbelldenied "github.com/niflaot/pixels/networking/outbound/room/doorbell/denied"
	outdoorbellhide "github.com/niflaot/pixels/networking/outbound/room/doorbell/hide"
)

// TestCommandName verifies the stable command name.
func TestCommandName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
}

// TestHandleRejectsResponderOutsideRoom verifies room presence gating.
func TestHandleRejectsResponderOutsideRoom(t *testing.T) {
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("owner", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "Owner"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "owner", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}
	handler := Handler{Players: players, Bindings: bindings}
	err = handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{
		Handler: netconn.Context{ConnectionID: "owner", ConnectionKind: "websocket"}, Username: "Guest",
	}})
	if !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("expected not in room, got %v", err)
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter missing active room: %v", err)
	}
	handler.Runtime = roomlive.NewRegistry(nil)
	err = handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: netconn.Context{ConnectionID: "owner", ConnectionKind: "websocket"}}})
	if !errors.Is(err, ErrNotInRoom) {
		t.Fatalf("expected missing active room, got %v", err)
	}
}

// TestHandleRejectsWaitingPlayer verifies both participants receive rejection state.
func TestHandleRejectsWaitingPlayer(t *testing.T) {
	fixture := newCommandFixture(t)
	err := fixture.handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{
		Handler: fixture.ownerContext, Username: "Guest", Accepted: false,
	}})
	if err != nil {
		t.Fatalf("reject doorbell: %v", err)
	}
	if fixture.active.DoorbellLen() != 0 || lastHeader(fixture.ownerPackets) != outdoorbelldenied.Header || lastHeader(fixture.guestPackets) != outdoorbelldenied.Header {
		t.Fatalf("unexpected rejection owner=%#v guest=%#v", *fixture.ownerPackets, *fixture.guestPackets)
	}
}

// TestHandleAcceptsWaitingPlayer verifies acceptance is emitted before normal entry.
func TestHandleAcceptsWaitingPlayer(t *testing.T) {
	fixture := newCommandFixture(t)
	err := fixture.handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{
		Handler: fixture.ownerContext, Username: "Guest", Accepted: true,
	}})
	if err == nil {
		t.Fatal("expected zero enter handler dependencies to reject final join")
	}
	if fixture.active.DoorbellLen() != 0 || lastHeader(fixture.ownerPackets) != outdoorbellhide.Header || lastHeader(fixture.guestPackets) != outdoorbellhide.Header {
		t.Fatalf("unexpected acceptance owner=%#v guest=%#v", *fixture.ownerPackets, *fixture.guestPackets)
	}
}

// TestHandleRejectsMissingRequest verifies stale owner decisions are ignored safely.
func TestHandleRejectsMissingRequest(t *testing.T) {
	fixture := newCommandFixture(t)
	if _, found := fixture.active.ResolveDoorbell("Guest"); !found {
		t.Fatal("expected queued guest")
	}
	err := fixture.handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{
		Handler: fixture.ownerContext, Username: "Guest", Accepted: false,
	}})
	if !errors.Is(err, ErrRequestNotFound) {
		t.Fatalf("expected missing request, got %v", err)
	}
}

// TestHandleToleratesMissingOwnerConnection verifies waiter rejection remains deliverable.
func TestHandleToleratesMissingOwnerConnection(t *testing.T) {
	fixture := newCommandFixture(t)
	fixture.handler.Connections.Remove("websocket", "owner")
	err := fixture.handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{
		Handler: fixture.ownerContext, Username: "Guest", Accepted: false,
	}})
	if err != nil || lastHeader(fixture.guestPackets) != outdoorbelldenied.Header {
		t.Fatalf("unexpected missing-owner result packets=%#v err=%v", *fixture.guestPackets, err)
	}
}

// TestHandleRejectsNonRightsResponder verifies room-scoped approval authorization.
func TestHandleRejectsNonRightsResponder(t *testing.T) {
	fixture := newCommandFixture(t)
	fixture.bindings.RemoveByPlayer(7)
	peer, err := playerlive.NewSessionPeer("owner", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create responder peer: %v", err)
	}
	responder, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 6, Username: "Visitor"}, peer)
	if err != nil {
		t.Fatalf("create responder: %v", err)
	}
	if err := responder.EnterRoom(9); err != nil {
		t.Fatalf("enter responder room: %v", err)
	}
	if err := fixture.players.Add(responder); err != nil {
		t.Fatalf("add responder: %v", err)
	}
	if err := fixture.bindings.Add(binding.Binding{PlayerID: 6, ConnectionID: "owner", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("bind responder: %v", err)
	}
	err = fixture.handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{
		Handler: fixture.ownerContext, Username: "Guest", Accepted: true,
	}})
	if !errors.Is(err, roomentry.ErrAccessDenied) || fixture.active.DoorbellLen() != 1 {
		t.Fatalf("expected access denied, got %v", err)
	}
}

// commandFixture stores one active doorbell command setup.
type commandFixture struct {
	// handler stores command behavior.
	handler Handler
	// active stores the active room.
	active *roomlive.Room
	// ownerContext stores the responder connection context.
	ownerContext netconn.Context
	// ownerPackets stores packets sent to the responder.
	ownerPackets *[]codec.Packet
	// guestPackets stores packets sent to the waiting player.
	guestPackets *[]codec.Packet
	// players stores live players.
	players *playerlive.Registry
	// bindings stores player connections.
	bindings *binding.Registry
}

// newCommandFixture creates one owner and waiting guest setup.
func newCommandFixture(t *testing.T) commandFixture {
	t.Helper()
	ownerContext, ownerSession, ownerPackets := commandContext(t, "owner")
	guestContext, _, guestPackets := commandContext(t, "guest")
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("owner", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	owner, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "Owner"}, peer)
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if err := owner.EnterRoom(9); err != nil {
		t.Fatalf("enter owner room: %v", err)
	}
	if err := players.Add(owner); err != nil {
		t.Fatalf("add owner: %v", err)
	}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "owner", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("bind owner: %v", err)
	}
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 7, Username: "Owner", ConnectionID: "owner", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join owner: %v", err)
	}
	if !active.RequestDoorbell(roomdoorbell.Entry{PlayerID: 8, Username: "Guest", Handler: guestContext, RequestedAt: time.Now()}, true) {
		t.Fatal("queue guest")
	}
	connections := netconn.NewRegistry()
	if err := connections.Register(ownerSession); err != nil {
		t.Fatalf("register owner connection: %v", err)
	}
	entry := roomentry.New(roomentry.Config{}, nil, nil, nil, roomentry.Nodes{})

	return commandFixture{
		handler: Handler{Players: players, Bindings: bindings, Runtime: runtime, Connections: connections, Entry: entry},
		active:  active, ownerContext: ownerContext, ownerPackets: ownerPackets, guestPackets: guestPackets,
		players: players, bindings: bindings,
	}
}
