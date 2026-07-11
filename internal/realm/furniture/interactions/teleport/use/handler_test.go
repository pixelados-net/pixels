package use

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	teleport "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inuse "github.com/niflaot/pixels/networking/inbound/furniture/use"
)

// TestHandleRoutesBoundPlayerAndSoftlyIgnoresMissingFurniture verifies command behavior.
func TestHandleRoutesBoundPlayerAndSoftlyIgnoresMissingFurniture(t *testing.T) {
	handler, connection := handlerForUseTest(t)
	cmd := Command{Handler: connection, ItemID: 99, State: 1}
	if cmd.CommandName() != Name {
		t.Fatalf("unexpected command name %q", cmd.CommandName())
	}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: cmd}); err != nil {
		t.Fatalf("handle missing furniture: %v", err)
	}
}

// TestPacketAdapterDecodesAndDispatches verifies handler registration and adaptation.
func TestPacketAdapterDecodesAndDispatches(t *testing.T) {
	handler, connection := handlerForUseTest(t)
	registry := netconn.NewHandlerRegistry()
	Register(registry, New(handler, nil))
	packet, err := codec.NewPacket(inuse.Header, inuse.Definition, codec.Int32(99), codec.Int32(1))
	if err != nil {
		t.Fatalf("create use packet: %v", err)
	}
	connection.State = netconn.StateConnected
	connection.Authenticated = true
	if err := registry.Handle(connection, packet); err != nil {
		t.Fatalf("route use packet: %v", err)
	}
}

// TestHandleRejectsMissingBinding verifies session guards.
func TestHandleRejectsMissingBinding(t *testing.T) {
	err := (Handler{}).Handle(context.Background(), command.Envelope[Command]{Command: Command{}})
	if err == nil {
		t.Fatal("expected missing binding error")
	}
}

// handlerForUseTest creates a bound live player and active room.
func handlerForUseTest(t *testing.T) (Handler, netconn.Context) {
	t.Helper()
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("one", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create session peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room presence: %v", err)
	}
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "one", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}
	runtime := roomlive.NewRegistry(nil)
	if _, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25}); err != nil {
		t.Fatalf("activate room: %v", err)
	}
	teleports := teleport.NewService(teleport.Config{}, nil, runtime, nil, nil, nil)
	connection := netconn.Context{ConnectionID: "one", ConnectionKind: "websocket"}

	return Handler{Players: players, Bindings: bindings, Runtime: runtime, Teleports: teleports}, connection
}
