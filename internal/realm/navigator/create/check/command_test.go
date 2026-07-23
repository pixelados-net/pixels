package cancreate

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcancreate "github.com/niflaot/pixels/networking/outbound/navigator/create/cancreate"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ownerListerForTest returns a configured room count.
type ownerListerForTest struct {
	// count stores returned room count.
	count int
	// err stores a read failure.
	err error
}

// ListByOwner returns a bounded room fixture list.
func (lister ownerListerForTest) ListByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return make([]roommodel.Room, lister.count), lister.err
}

// TestCommandLoggingWritesConnection verifies command diagnostics.
func TestCommandLoggingWritesConnection(t *testing.T) {
	var output bytes.Buffer
	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.DebugLevel))
	logger.Debug("command", zap.Object("command", Command{Handler: netconn.Context{ConnectionID: "conn"}}))
	if !strings.Contains(output.String(), "conn") {
		t.Fatalf("log=%s", output.String())
	}
}

// TestCommandName verifies the command name.
func TestCommandName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name")
	}
}

// TestHandleProjectsCurrentRoomLimit verifies native preflight responses.
func TestHandleProjectsCurrentRoomLimit(t *testing.T) {
	for _, count := range []int{0, roomservice.MaxRoomsPerPlayer} {
		players, bindings, connection, packets := canCreateFixture(t)
		handler := Handler{Players: players, Bindings: bindings, Rooms: ownerListerForTest{count: count}}
		if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection}}); err != nil {
			t.Fatalf("count=%d handle: %v", count, err)
		}
		values, err := codec.DecodePacketExact((*packets)[0], outcancreate.Definition)
		expected := ResultAllowed
		if count == roomservice.MaxRoomsPerPlayer {
			expected = ResultLimitReached
		}
		if err != nil || values[0].Int32 != expected || values[1].Int32 != RoomLimit {
			t.Fatalf("count=%d values=%+v err=%v", count, values, err)
		}
	}
}

// TestHandlePropagatesRoomReadFailures verifies persistence errors remain explicit.
func TestHandlePropagatesRoomReadFailures(t *testing.T) {
	players, bindings, connection, _ := canCreateFixture(t)
	cause := errors.New("database unavailable")
	handler := Handler{Players: players, Bindings: bindings, Rooms: ownerListerForTest{err: cause}}
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection}})
	if !errors.Is(err, cause) {
		t.Fatalf("error=%v", err)
	}
}

// canCreateFixture creates one bound player and packet-capturing connection.
func canCreateFixture(t *testing.T) (*playerlive.Registry, *binding.Registry, netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	_ = inbound.Register(1, func(ctx netconn.Context, _ codec.Packet) error { connection = ctx; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	packets := make([]codec.Packet, 0, 1)
	session, _ := netconn.NewSession(netconn.SessionConfig{ID: "can-create", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	_ = session.Receive(context.Background(), codec.Packet{Header: 1})
	peer, _ := playerlive.NewSessionPeer("can-create", "websocket", time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	players := playerlive.NewRegistry()
	_ = players.Add(player)
	bindings := binding.NewRegistry()
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "can-create", ConnectionKind: "websocket"})

	return players, bindings, connection, &packets
}
