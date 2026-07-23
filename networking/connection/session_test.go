package connection

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/networking/codec"
)

// TestSessionReceiveRoutesPacket verifies inbound handling.
func TestSessionReceiveRoutesPacket(t *testing.T) {
	session := mustSession(t, sessionFixture(t))
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive packet: %v", err)
	}

	if session.State() != StateHandshaking {
		t.Fatalf("expected handshaking state, got %d", session.State())
	}
}

// TestSessionSendRoutesAndWrites verifies outbound handling and sending.
func TestSessionSendRoutesAndWrites(t *testing.T) {
	fixture := sessionFixture(t)
	session := mustSession(t, fixture)
	if err := session.Send(context.Background(), codec.Packet{Header: 2}); err != nil {
		t.Fatalf("send packet: %v", err)
	}

	if *fixture.sent != 1 {
		t.Fatalf("expected %d sent packets, got %d", 1, *fixture.sent)
	}
}

// TestSessionPacketLoggerRecordsTraffic verifies packet logger hooks.
func TestSessionPacketLoggerRecordsTraffic(t *testing.T) {
	fixture := sessionFixture(t)
	logger := &testPacketLogger{}
	fixture.PacketLogger = logger
	session := mustSession(t, fixture)

	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive packet: %v", err)
	}
	if err := session.Send(context.Background(), codec.Packet{Header: 2}); err != nil {
		t.Fatalf("send packet: %v", err)
	}

	if logger.received != 1 || logger.sent != 1 || logger.unhandled != 0 {
		t.Fatalf("unexpected logger counters: %#v", logger)
	}
}

// TestSessionPacketLoggerRecordsUnhandled verifies missing inbound handlers are warned.
func TestSessionPacketLoggerRecordsUnhandled(t *testing.T) {
	fixture := sessionFixture(t)
	logger := &testPacketLogger{}
	fixture.PacketLogger = logger
	session := mustSession(t, fixture)

	err := session.Receive(context.Background(), codec.Packet{Header: 99})
	if err != nil {
		t.Fatalf("expected ignored missing handler, got %v", err)
	}

	if logger.received != 1 || logger.unhandled != 1 {
		t.Fatalf("unexpected logger counters: %#v", logger)
	}
}

// TestSessionPacketLoggerRecordsMultipleUnhandled verifies unknown packets keep flowing.
func TestSessionPacketLoggerRecordsMultipleUnhandled(t *testing.T) {
	fixture := sessionFixture(t)
	logger := &testPacketLogger{}
	fixture.PacketLogger = logger
	session := mustSession(t, fixture)

	for _, header := range []uint16{99, 100, 101} {
		if err := session.Receive(context.Background(), codec.Packet{Header: header}); err != nil {
			t.Fatalf("expected ignored missing handler, got %v", err)
		}
	}

	if logger.received != 3 || logger.unhandled != 3 {
		t.Fatalf("unexpected logger counters: %#v", logger)
	}
}

// TestSessionAuthenticateTracksTime verifies authentication state.
func TestSessionAuthenticateTracksTime(t *testing.T) {
	session := mustSession(t, sessionFixture(t))
	authenticatedAt := time.Unix(20, 0)
	mustTransition(t, session, EventPacketReceived)
	mustTransition(t, session, EventAuthenticationStarted)
	if err := session.Authenticate(authenticatedAt); err != nil {
		t.Fatalf("authenticate session: %v", err)
	}

	got, ok := session.AuthenticatedAt()
	if !ok {
		t.Fatal("expected authenticated session")
	}

	if !got.Equal(authenticatedAt) {
		t.Fatalf("expected %s, got %s", authenticatedAt, got)
	}

	if session.State() != StateAuthenticated {
		t.Fatalf("expected authenticated state, got %d", session.State())
	}
}

// TestSessionDisconnectDisposesOnce verifies disposal behavior.
func TestSessionDisconnectDisposesOnce(t *testing.T) {
	fixture := sessionFixture(t)
	session := mustSession(t, fixture)
	reason := Reason{Code: DisconnectLocalClose}

	if err := session.Disconnect(context.Background(), reason); err != nil {
		t.Fatalf("disconnect session: %v", err)
	}

	select {
	case <-session.Done():
	default:
		t.Fatal("expected done channel closed")
	}

	if *fixture.disposed != 1 {
		t.Fatalf("expected %d disposals, got %d", 1, *fixture.disposed)
	}

	if err := session.Disconnect(context.Background(), reason); !errors.Is(err, ErrDisposed) {
		t.Fatalf("expected disposed error, got %v", err)
	}
}

// TestSessionDisconnectClosesSecurityAfterDisposal verifies terminal writes can remain encrypted.
func TestSessionDisconnectClosesSecurityAfterDisposal(t *testing.T) {
	fixture := sessionFixture(t)
	channel := &fakeSecureChannel{state: SecurityReady}
	securityReadyDuringDispose := false
	fixture.Disposer = func(context.Context, Reason) error {
		securityReadyDuringDispose = channel.State() == SecurityReady
		return nil
	}
	session := mustSession(t, fixture)
	if err := session.AttachSecurity(channel); err != nil {
		t.Fatalf("attach security: %v", err)
	}

	if err := session.Disconnect(context.Background(), Reason{Code: DisconnectKicked}); err != nil {
		t.Fatalf("disconnect session: %v", err)
	}
	if !securityReadyDuringDispose {
		t.Fatal("expected security to remain ready during transport disposal")
	}
	if channel.State() != SecurityFailed {
		t.Fatalf("expected security closed after disposal, got %d", channel.State())
	}
}

// TestSessionRejectsInvalidConfig verifies required transport callbacks.
func TestSessionRejectsInvalidConfig(t *testing.T) {
	_, err := NewSession(SessionConfig{})
	if !errors.Is(err, ErrInvalidConnectionConfig) {
		t.Fatalf("expected invalid config, got %v", err)
	}
}

// TestSessionRejectsAfterDisconnect verifies disposed operation protection.
func TestSessionRejectsAfterDisconnect(t *testing.T) {
	session := mustSession(t, sessionFixture(t))
	if err := session.Disconnect(context.Background(), UnknownReason()); err != nil {
		t.Fatalf("disconnect session: %v", err)
	}

	err := session.Receive(context.Background(), codec.Packet{Header: 1})
	if !errors.Is(err, ErrDisposed) {
		t.Fatalf("expected disposed receive, got %v", err)
	}

	err = session.Send(context.Background(), codec.Packet{Header: 2})
	if !errors.Is(err, ErrDisposed) {
		t.Fatalf("expected disposed send, got %v", err)
	}

	if err := session.Authenticate(time.Now()); !errors.Is(err, ErrDisposed) {
		t.Fatalf("expected disposed authenticate, got %v", err)
	}
}

// sessionFixtureConfig extends session config with counters.
type sessionFixtureConfig struct {
	// SessionConfig stores the base connection fixture.
	SessionConfig
	// sent counts outbound packets.
	sent *int
	// disposed counts disposal calls.
	disposed *int
}

// sessionFixture returns a configured test session fixture.
func sessionFixture(t *testing.T) sessionFixtureConfig {
	t.Helper()
	inbound := NewHandlerRegistry()
	outbound := NewHandlerRegistry()
	sent := 0
	disposed := 0

	mustRegister(t, inbound, 1, "inbound")
	mustRegister(t, outbound, 2, "outbound")

	return sessionFixtureConfig{
		SessionConfig: SessionConfig{
			ID:        "one",
			Kind:      "websocket",
			StartedAt: time.Unix(10, 0),
			Inbound:   inbound,
			Outbound:  outbound,
			Sender: func(context.Context, codec.Packet) error {
				sent++
				return nil
			},
			Disposer: func(context.Context, Reason) error {
				disposed++
				return nil
			},
		},
		sent:     &sent,
		disposed: &disposed,
	}
}

// mustRegister registers a packet handler or fails the test.
func mustRegister(t *testing.T, registry *HandlerRegistry, header uint16, name string) {
	t.Helper()
	handler := func(context Context, packet codec.Packet) error {
		if context.Direction == 0 || context.ConnectionID == "" || packet.Header != header {
			t.Fatalf("invalid handler context for %s", name)
		}

		return nil
	}

	if err := registry.Register(header, handler, AllowAnyActiveState(), AllowUnauthenticated()); err != nil {
		t.Fatalf("register handler: %v", err)
	}
}

// mustSession creates a session or fails the test.
func mustSession(t *testing.T, config sessionFixtureConfig) *Session {
	t.Helper()
	session, err := NewSession(config.SessionConfig)
	if err != nil {
		t.Fatalf("new session: %v", err)
	}

	return session
}

// testPacketLogger records packet logger calls.
type testPacketLogger struct {
	// received, sent, and unhandled count packet logger calls.
	received, sent, unhandled int
}

// Received records an inbound packet.
func (logger *testPacketLogger) Received(Context, codec.Packet) {
	logger.received++
}

// Sent records an outbound packet.
func (logger *testPacketLogger) Sent(Context, codec.Packet) {
	logger.sent++
}

// Unhandled records a packet without a handler.
func (logger *testPacketLogger) Unhandled(Context, codec.Packet) {
	logger.unhandled++
}

// mustTransition applies a transition or fails the test.
func mustTransition(t *testing.T, session *Session, event Event) {
	t.Helper()
	if err := session.Transition(event); err != nil {
		t.Fatalf("transition %s: %v", event, err)
	}
}
