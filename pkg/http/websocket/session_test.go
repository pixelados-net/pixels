package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestSocketSessionSendQueue verifies bounded outbound queue behavior.
func TestSocketSessionSendQueue(t *testing.T) {
	socket := testQueuedSocket(t, 1)

	if err := socket.send(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("send packet: %v", err)
	}
	if err := socket.send(context.Background(), codec.Packet{Header: 2}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected queue full, got %v", err)
	}

	socket.closed.Store(true)
	if err := socket.send(context.Background(), codec.Packet{Header: 3}); !errors.Is(err, netconn.ErrDisposed) {
		t.Fatalf("expected disposed, got %v", err)
	}
}

// TestSocketSessionActivateQueue verifies queued security activation.
func TestSocketSessionActivateQueue(t *testing.T) {
	socket := testQueuedSocket(t, 1)

	if err := socket.activate(context.Background(), testSecureChannel{}); err != nil {
		t.Fatalf("activate security: %v", err)
	}

	item := <-socket.queue
	if item.kind != writeActivate {
		t.Fatalf("expected activate item, got %#v", item)
	}
}

// TestSocketSessionActivateBackpressure verifies activation queue errors.
func TestSocketSessionActivateBackpressure(t *testing.T) {
	socket := testQueuedSocket(t, 1)
	socket.queue <- writeItem{kind: writePacket}

	if err := socket.activate(context.Background(), testSecureChannel{}); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected queue full, got %v", err)
	}

	socket.closed.Store(true)
	if err := socket.activate(context.Background(), testSecureChannel{}); !errors.Is(err, netconn.ErrDisposed) {
		t.Fatalf("expected disposed, got %v", err)
	}
}

// TestSocketSessionDisposeClosesOnce verifies graceful close enqueueing.
func TestSocketSessionDisposeClosesOnce(t *testing.T) {
	socket := testQueuedSocket(t, 4)

	err := socket.dispose(context.Background(), netconn.Reason{Code: netconn.DisconnectProtocolError, Message: "bad"})
	if err != nil {
		t.Fatalf("dispose: %v", err)
	}
	if len(socket.queue) != 3 {
		t.Fatalf("expected protocol reason, error, and close frames, got %d", len(socket.queue))
	}
	if err := socket.dispose(context.Background(), netconn.Reason{}); !errors.Is(err, netconn.ErrDisposed) {
		t.Fatalf("expected disposed, got %v", err)
	}
}

// TestSocketSessionWaitFinishes verifies graceful wait timeout cleanup.
func TestSocketSessionWaitFinishes(t *testing.T) {
	socket := testQueuedSocket(t, 1)
	socket.config.CloseGrace = time.Nanosecond
	socket.wait()

	select {
	case <-socket.done:
	default:
		t.Fatal("expected done to close")
	}
}

// TestPacketLoggerForEnvironmentRecordsDevelopment verifies development packet logs.
func TestPacketLoggerForEnvironmentRecordsDevelopment(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	trafficLogger := packetLoggerForEnvironment("development", zap.New(core), logger.Config{})
	if trafficLogger == nil {
		t.Fatal("expected development packet logger")
	}

	context := netconn.Context{ConnectionID: "one", ConnectionKind: kind, State: netconn.StateConnected}
	packet := codec.Packet{Header: 7, Payload: []byte{1, 2}}

	trafficLogger.Received(context, packet)
	trafficLogger.Sent(context, packet)
	trafficLogger.Unhandled(context, packet)
	trafficLogger.(packetLogger).Disconnected(context, netconn.Reason{Code: netconn.DisconnectAuthenticationFailed, Message: "sso ticket not found"})

	if logs.Len() != 4 {
		t.Fatalf("expected four packet logs, got %d", logs.Len())
	}
	if logs.All()[0].Message != "packet received" || logs.All()[1].Message != "packet sent" {
		t.Fatalf("unexpected packet log messages: %#v", logs.All())
	}
	if logs.All()[2].Level != zap.WarnLevel || logs.All()[2].Message != "packet unhandled" {
		t.Fatalf("unexpected unhandled packet log: %#v", logs.All()[2])
	}
	if logs.All()[3].Message != "websocket disconnecting" {
		t.Fatalf("unexpected disconnect log: %#v", logs.All()[3])
	}
}

// TestPacketLoggerForEnvironmentRecordsToonFields verifies toon packet logs are compact.
func TestPacketLoggerForEnvironmentRecordsToonFields(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	trafficLogger := packetLoggerForEnvironment("development", zap.New(core), logger.Config{ToonConsole: true})
	if trafficLogger == nil {
		t.Fatal("expected development packet logger")
	}

	context := netconn.Context{
		ConnectionID:   "12345678-1234-1234-1234-123456789012",
		ConnectionKind: kind,
		State:          netconn.StateConnected,
	}

	trafficLogger.Received(context, codec.Packet{Header: 4000, Payload: []byte{1}})

	fields := logs.All()[0].ContextMap()
	if fields["cid"] != "12345678" {
		t.Fatalf("expected short connection id, got %#v", fields)
	}
	if fields["connection_kind"] != nil {
		t.Fatalf("expected no connection kind, got %#v", fields)
	}
	if fields["header"] != uint64(4000) && fields["header"] != uint16(4000) {
		t.Fatalf("expected compact header field, got %#v", fields["header"])
	}
}

// TestPacketLoggerForEnvironmentSkipsProduction verifies production packet logs are disabled.
func TestPacketLoggerForEnvironmentSkipsProduction(t *testing.T) {
	if packetLoggerForEnvironment("production", zap.NewNop(), logger.Config{}) != nil {
		t.Fatal("expected production packet logger to be disabled")
	}
}

// BenchmarkSocketSessionSend measures outbound queue enqueue overhead.
func BenchmarkSocketSessionSend(b *testing.B) {
	socket := testQueuedSocket(b, 1)
	packet := codec.Packet{Header: 1}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := socket.send(ctx, packet); err != nil {
			b.Fatalf("send packet: %v", err)
		}
		<-socket.queue
	}
}

// testQueuedSocket creates a socket with a bounded write queue.
func testQueuedSocket(t testing.TB, size int) *socketSession {
	t.Helper()

	return &socketSession{
		config:  Config{QueueSize: size, CloseGrace: time.Millisecond}.Normalize(),
		queue:   make(chan writeItem, size),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
		session: testWebSocketSession(t, netconn.NewHandlerRegistry()),
		log:     zap.NewNop(),
	}
}

// testWebSocketSession creates a base connection session.
func testWebSocketSession(t testing.TB, inbound *netconn.HandlerRegistry) *netconn.Session {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error {
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       "socket-test",
		Kind:     kind,
		Inbound:  inbound,
		Outbound: outbound,
		Sender: func(context.Context, codec.Packet) error {
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	return session
}
