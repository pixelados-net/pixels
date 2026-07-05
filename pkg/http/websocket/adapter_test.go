package websocket

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	fastws "github.com/fasthttp/websocket"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrelease "github.com/niflaot/pixels/networking/inbound/handshake/release"
	inticket "github.com/niflaot/pixels/networking/inbound/security/ticket"
	outauth "github.com/niflaot/pixels/networking/outbound/authentication/ok"
	appconfig "github.com/niflaot/pixels/pkg/config/app"
	"github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/zap"
)

// TestAdapterAuthenticatesWebSocket verifies release and SSO over a real socket.
func TestAdapterAuthenticatesWebSocket(t *testing.T) {
	service := testService(t)
	ticket, err := service.Create(context.Background(), sso.CreateRequest{UserID: "todo-user", TTL: time.Minute})
	if err != nil {
		t.Fatalf("create ticket: %v", err)
	}

	url, shutdown := testServer(t, service)
	defer shutdown()

	conn, _, err := fastws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close()

	writeClientPacket(t, conn, releasePacket(t))
	writeClientPacket(t, conn, ticketPacket(t, ticket.Value))

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read authenticated: %v", err)
	}

	packets, _, err := codec.DecodeFrames(nil, data)
	if err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(packets) != 1 || packets[0].Header != outauth.Header {
		t.Fatalf("expected authenticated packet, got %#v", packets)
	}
}

// testServer starts a Fiber WebSocket server.
func testServer(t *testing.T, service *sso.Service) (string, func()) {
	t.Helper()
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	adapter := New(
		Config{PingInterval: time.Hour, PongTimeout: time.Hour},
		appconfig.Config{Environment: "development"},
		netconn.NewRegistry(),
		realmconn.NewHandlers(service),
		zap.NewNop(),
	)
	app.Get("/ws", websocket.New(adapter.Handle))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	go func() {
		_ = app.Listener(listener)
	}()

	return "ws://" + listener.Addr().String() + "/ws", func() {
		_ = app.Shutdown()
	}
}

// writeClientPacket writes one protocol packet.
func writeClientPacket(t *testing.T, conn *fastws.Conn, packet codec.Packet) {
	t.Helper()
	frame, err := codec.AppendFrame(nil, packet)
	if err != nil {
		t.Fatalf("append frame: %v", err)
	}
	if err := conn.WriteMessage(fastws.BinaryMessage, frame); err != nil {
		t.Fatalf("write packet: %v", err)
	}
}

// releasePacket creates a release packet.
func releasePacket(t *testing.T) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(
		inrelease.Header,
		inrelease.Definition,
		codec.String("NITRO-test"),
		codec.String("HTML5"),
		codec.Int32(0),
		codec.Int32(0),
	)
	if err != nil {
		t.Fatalf("release packet: %v", err)
	}

	return packet
}

// ticketPacket creates an SSO ticket packet.
func ticketPacket(t *testing.T, ticket string) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(inticket.Header, inticket.Definition, codec.String(ticket), codec.Int32(1))
	if err != nil {
		t.Fatalf("ticket packet: %v", err)
	}

	return packet
}

// testService creates an SSO service.
func testService(t *testing.T) *sso.Service {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.New(redis.Config{Address: server.Addr()})
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close redis: %v", err)
		}
	})

	return sso.New(sso.Config{DefaultTTL: time.Minute, Key: "test-key", Prefix: "pixels:sso"}, client)
}
