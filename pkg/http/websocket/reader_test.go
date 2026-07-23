package websocket

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestReceiveBuffersPartialFrames verifies fragmented transport messages.
func TestReceiveBuffersPartialFrames(t *testing.T) {
	socket, received := testReaderSocket(t)
	frame := testFrame(t, codec.Packet{Header: 77})

	if reason, ok := socket.receive(context.Background(), frame[:3]); !ok {
		t.Fatalf("partial frame failed: %#v", reason)
	}
	if *received != 0 {
		t.Fatalf("expected no complete packet, got %d", *received)
	}

	if reason, ok := socket.receive(context.Background(), frame[3:]); !ok {
		t.Fatalf("complete frame failed: %#v", reason)
	}
	if *received != 1 {
		t.Fatalf("expected one packet, got %d", *received)
	}
}

// TestReceiveRejectsInvalidFrames verifies protocol framing errors.
func TestReceiveRejectsInvalidFrames(t *testing.T) {
	socket, _ := testReaderSocket(t)
	_, ok := socket.receive(context.Background(), []byte{0, 0, 0, 1, 0})
	if ok {
		t.Fatal("expected invalid frame to close")
	}
}

// TestReceiveIgnoresMultipleMissingHandlers verifies unknown frame packets continue.
func TestReceiveIgnoresMultipleMissingHandlers(t *testing.T) {
	socket, _ := testReaderSocket(t)
	frame := append(testFrame(t, codec.Packet{Header: 99}), testFrame(t, codec.Packet{Header: 100})...)

	if reason, ok := socket.receive(context.Background(), frame); !ok {
		t.Fatalf("expected receive to continue, got %#v", reason)
	}
}

// TestDispatchIgnoresMissingHandlers verifies unknown packets stay non-fatal.
func TestDispatchIgnoresMissingHandlers(t *testing.T) {
	socket, _ := testReaderSocket(t)
	reason, ok := socket.dispatch(context.Background(), codec.Packet{Header: 99})
	if !ok {
		t.Fatalf("expected dispatch to continue, got %#v", reason)
	}
}

// BenchmarkSocketSessionReceive measures frame buffering, decoding, and dispatch.
func BenchmarkSocketSessionReceive(b *testing.B) {
	socket, _ := testReaderSocket(b)
	frame := testFrame(b, codec.Packet{Header: 77})
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if reason, ok := socket.receive(ctx, frame); !ok {
			b.Fatalf("receive failed: %#v", reason)
		}
	}
}

// testReaderSocket creates a socket with one inbound handler.
func testReaderSocket(t testing.TB) (*socketSession, *int) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	received := 0
	if err := inbound.Register(77, func(netconn.Context, codec.Packet) error {
		received++
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	session := testWebSocketSession(t, inbound)

	return &socketSession{session: session}, &received
}

// testFrame encodes one packet frame.
func testFrame(t testing.TB, packet codec.Packet) []byte {
	t.Helper()
	frame, err := codec.AppendFrame(nil, packet)
	if err != nil {
		t.Fatalf("append frame: %v", err)
	}

	return frame
}
