package player

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.uber.org/zap"
)

// permissionChecker returns one configured permission decision.
type permissionChecker struct {
	// allowed stores the fixture result.
	allowed bool
}

// HasPermission returns the fixture permission decision.
func (checker permissionChecker) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return checker.allowed, nil
}

// TestAccessReadsAndActsOnLivePlayers verifies bounded player capabilities.
func TestAccessReadsAndActsOnLivePlayers(t *testing.T) {
	access, sent, disconnected := playerFixture(t)
	players := access.All()
	if len(players) != 1 || players[0].Username != "alice" || players[0].RoomID != 9 {
		t.Fatalf("unexpected snapshots: %+v", players)
	}
	if found, ok := access.Find(1); !ok || found.ID != 1 {
		t.Fatalf("expected player lookup, player=%+v found=%v", found, ok)
	}
	if _, ok := access.Find(99); ok {
		t.Fatal("expected missing player")
	}
	allowed, err := access.HasPermission(1, "plugin.test.use")
	if err != nil || !allowed {
		t.Fatalf("expected permission, allowed=%v err=%v", allowed, err)
	}
	if _, err = access.HasPermission(1, "*"); !errors.Is(err, permission.ErrInvalidRegistration) {
		t.Fatalf("expected invalid permission rejection, got %v", err)
	}
	if err = access.Message(1, "hello"); err != nil || len(*sent) != 1 {
		t.Fatalf("expected one alert, sent=%d err=%v", len(*sent), err)
	}
	if err = access.Message(99, "missing"); !errors.Is(err, binding.ErrBindingNotFound) {
		t.Fatalf("expected missing message target, got %v", err)
	}
	if err = access.Disconnect(99, "missing"); !errors.Is(err, binding.ErrBindingNotFound) {
		t.Fatalf("expected missing disconnect target, got %v", err)
	}
	if err = access.Disconnect(1, "qa"); err != nil || !*disconnected {
		t.Fatalf("expected disconnect, disconnected=%v err=%v", *disconnected, err)
	}
}

// TestAccessRegistersAuthenticatedInterceptor verifies SDK packet projection and flow.
func TestAccessRegistersAuthenticatedInterceptor(t *testing.T) {
	access, _, _ := playerFixture(t)
	header := uint16(44)
	nativeCalls := 0
	if err := access.inbound.Register(header, func(netconn.Context, codec.Packet) error {
		nativeCalls++
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register native handler: %v", err)
	}
	observed := sdkplugin.InterceptContext{}
	if err := access.Intercept(nil, sdkplugin.InterceptOptions{}); !errors.Is(err, netconn.ErrInvalidHandler) {
		t.Fatalf("expected nil interceptor rejection, got %v", err)
	}
	if err := access.Intercept(func(ctx context.Context, packet sdkplugin.InterceptContext, next sdkplugin.Next) error {
		observed = packet
		return next(ctx)
	}, sdkplugin.InterceptOptions{Header: &header, Priority: sdkplugin.PriorityHigh}); err != nil {
		t.Fatalf("register interceptor: %v", err)
	}
	err := access.inbound.Handle(netconn.Context{ConnectionID: "one", ConnectionKind: "websocket", State: netconn.StateCreated}, codec.Packet{Header: header, Payload: []byte{1, 2}})
	if err != nil || nativeCalls != 1 || observed.Player.ID != 1 || observed.Header != header {
		t.Fatalf("expected intercepted native packet, observed=%+v native=%d err=%v", observed, nativeCalls, err)
	}
}

// playerFixture creates one fully bound live player and transport session.
func playerFixture(t *testing.T) (*Access, *[]codec.Packet, *bool) {
	t.Helper()
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	connections := netconn.NewRegistry()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 1)
	disconnected := false
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "one", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { disconnected = true; return nil },
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}
	peer, err := playerlive.NewSessionPeer("one", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	current, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 1, Username: "alice", Gender: playermodel.GenderFemale}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err = current.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	if err = players.Add(current); err != nil {
		t.Fatalf("add player: %v", err)
	}
	if err = bindings.Add(binding.Binding{PlayerID: 1, ConnectionID: "one", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}
	access := NewAccess(players, bindings, connections, inbound, permissionChecker{allowed: true}, time.Second, zap.NewNop(), pluginruntime.NewScope("test"))
	return access, &sent, &disconnected
}

// TestInterceptorPanicAdvancesNativeHandler verifies plugin failure does not lose the packet.
func TestInterceptorPanicAdvancesNativeHandler(t *testing.T) {
	scope := pluginruntime.NewScope("broken")
	access := &Access{timeout: time.Second, log: zap.NewNop(), scope: scope}
	nativeCalls := 0
	err := access.invokeInterceptor(context.Background(), sdkplugin.InterceptContext{Header: 7}, func(context.Context, sdkplugin.InterceptContext, sdkplugin.Next) error {
		panic("boom")
	}, func() error {
		nativeCalls++
		return nil
	})
	if err != nil || nativeCalls != 1 || scope.Enabled() {
		t.Fatalf("expected native continuation after panic, calls=%d enabled=%v err=%v", nativeCalls, scope.Enabled(), err)
	}
}

// TestInterceptorCanCancelPacket verifies omitting Next stops native dispatch.
func TestInterceptorCanCancelPacket(t *testing.T) {
	access := &Access{timeout: time.Second, log: zap.NewNop(), scope: pluginruntime.NewScope("cancel")}
	nativeCalls := 0
	err := access.invokeInterceptor(context.Background(), sdkplugin.InterceptContext{Header: 7}, func(context.Context, sdkplugin.InterceptContext, sdkplugin.Next) error {
		return nil
	}, func() error {
		nativeCalls++
		return nil
	})
	if err != nil || nativeCalls != 0 {
		t.Fatalf("expected cancellation before native dispatch, calls=%d err=%v", nativeCalls, err)
	}
}
