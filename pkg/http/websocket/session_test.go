package websocket

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/zap"
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
	socket := testQueuedSocket(t, 3)

	err := socket.dispose(context.Background(), netconn.Reason{Code: netconn.DisconnectProtocolError, Message: "bad"})
	if err != nil {
		t.Fatalf("dispose: %v", err)
	}
	if len(socket.queue) != 2 {
		t.Fatalf("expected protocol and close frames, got %d", len(socket.queue))
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
