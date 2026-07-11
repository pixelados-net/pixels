package votes

import (
	"context"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	votescmd "github.com/niflaot/pixels/internal/realm/room/control/commands/votes"
	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlike "github.com/niflaot/pixels/networking/inbound/room/like"
	"go.uber.org/zap"
)

// managerForTest records handler vote dispatch.
type managerForTest struct {
	// cast reports whether the command reached vote behavior.
	cast bool
}

// Cast records one vote.
func (manager *managerForTest) Cast(context.Context, int64, int64) (roomvotes.Mutation, error) {
	manager.cast = true
	return roomvotes.Mutation{Score: 1, Inserted: true}, nil
}

// State returns empty vote state.
func (*managerForTest) State(context.Context, int64, int64) (roomvotes.State, error) {
	return roomvotes.State{}, nil
}

// List returns no votes.
func (*managerForTest) List(context.Context, roomvotes.Query) ([]roomvotes.Vote, error) {
	return nil, nil
}

// TestNewCastDecodesAndDispatches verifies the packet adapter.
func TestNewCastDecodesAndDispatches(t *testing.T) {
	players, bindings, connection := handlerFixture(t)
	manager := &managerForTest{}
	handler := NewCast(votescmd.CastHandler{Players: players, Bindings: bindings, Votes: manager}, zap.NewNop())
	packet, err := codec.NewPacket(inlike.Header, inlike.Definition, codec.Int32(1))
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if err := handler(connection, packet); err != nil || !manager.cast {
		t.Fatalf("cast=%v err=%v", manager.cast, err)
	}
}

// TestNewCastRejectsWrongHeaders verifies packet header validation.
func TestNewCastRejectsWrongHeaders(t *testing.T) {
	handler := NewCast(votescmd.CastHandler{}, zap.NewNop())
	if err := handler(netconn.Context{}, codec.Packet{Header: inlike.Header + 1}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("error=%v", err)
	}
}

// TestRegisterCastRegistersProtocolHeader verifies registry wiring.
func TestRegisterCastRegistersProtocolHeader(t *testing.T) {
	registry := netconn.NewHandlerRegistry()
	RegisterCast(registry, func(netconn.Context, codec.Packet) error { return nil })
	context := netconn.Context{State: netconn.StateConnected, Authenticated: true}
	if err := registry.Handle(context, codec.Packet{Header: inlike.Header}); err != nil {
		t.Fatalf("handle registered packet: %v", err)
	}
}

// handlerFixture creates one bound room actor.
func handlerFixture(t *testing.T) (*playerlive.Registry, *binding.Registry, netconn.Context) {
	t.Helper()
	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 2, Username: "Alice"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	players := playerlive.NewRegistry()
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 2, ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	return players, bindings, netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}
}
