package connection

import (
	"context"
	"sync"
	"time"

	"github.com/niflaot/pixels/networking/codec"
)

// Session is a transport-agnostic connection implementation.
type Session struct {
	// mutex protects mutable session state.
	mutex sync.RWMutex
	// id identifies this session in its registry bucket.
	id ID
	// kind classifies the transport family.
	kind Kind
	// startedAt stores the session creation time.
	startedAt time.Time
	// remoteAddr stores the peer transport address.
	remoteAddr string
	// authenticatedAt stores the authentication completion time.
	authenticatedAt time.Time
	// lastPongAt stores the most recent heartbeat pong time.
	lastPongAt time.Time
	// authenticated reports whether authentication has completed.
	authenticated bool
	// disconnected reports whether disposal started.
	disconnected bool
	// disconnectReason stores the disposal reason.
	disconnectReason Reason
	// state stores the lifecycle phase.
	state State
	// trafficStarted prevents late security policy mutation.
	trafficStarted bool
	// policy controls whether security is required.
	policy SecurityPolicy
	// security opens and seals connection bytes.
	security SecureChannel
	// done closes when disposal starts.
	done chan struct{}
	// inbound routes received packets.
	inbound *HandlerRegistry
	// outbound routes sent packets.
	outbound *HandlerRegistry
	// sender writes outbound packets through the transport.
	sender Sender
	// disposer releases transport resources.
	disposer Disposer
	// activator installs security after queued writes.
	activator SecurityActivator
}

// NewSession creates a session connection.
func NewSession(config SessionConfig) (*Session, error) {
	if config.ID == "" || config.Kind == "" || config.Sender == nil || config.Disposer == nil {
		return nil, ErrInvalidConnectionConfig
	}

	startedAt := config.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now()
	}

	inbound := config.Inbound
	if inbound == nil {
		inbound = NewHandlerRegistry()
	}

	outbound := config.Outbound
	if outbound == nil {
		outbound = NewHandlerRegistry()
	}

	return &Session{
		id:         config.ID,
		kind:       config.Kind,
		startedAt:  startedAt,
		remoteAddr: config.RemoteAddr,
		lastPongAt: startedAt,
		state:      StateCreated,
		policy:     normalizeSecurityPolicy(config.SecurityPolicy),
		done:       make(chan struct{}),
		inbound:    inbound,
		outbound:   outbound,
		sender:     config.Sender,
		disposer:   config.Disposer,
		activator:  config.SecurityActivator,
	}, nil
}

// ID returns the connection identifier.
func (session *Session) ID() ID {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.id
}

// Kind returns the connection kind.
func (session *Session) Kind() Kind {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.kind
}

// StartedAt returns the connection start time.
func (session *Session) StartedAt() time.Time {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.startedAt
}

// RemoteAddr returns the transport peer address.
func (session *Session) RemoteAddr() string {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.remoteAddr
}

// AuthenticatedAt returns the authentication time when available.
func (session *Session) AuthenticatedAt() (time.Time, bool) {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.authenticatedAt, session.authenticated
}

// LastPongAt returns the last observed client pong time.
func (session *Session) LastPongAt() time.Time {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.lastPongAt
}

// MarkPong records a client heartbeat pong.
func (session *Session) MarkPong(at time.Time) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	if session.disconnected {
		return ErrDisposed
	}

	if at.IsZero() {
		at = time.Now()
	}

	session.lastPongAt = at

	return nil
}

// Authenticate marks the connection as authenticated.
func (session *Session) Authenticate(at time.Time) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	if session.disconnected {
		return ErrDisposed
	}

	if !canTransition(session.state, EventAuthenticationAccepted) {
		return ErrInvalidTransition
	}

	if at.IsZero() {
		at = time.Now()
	}

	session.authenticatedAt = at
	session.authenticated = true
	session.state = StateAuthenticated

	return nil
}

// Receive handles an inbound packet.
func (session *Session) Receive(ctx context.Context, packet codec.Packet) error {
	if err := session.markTraffic(EventPacketReceived); err != nil {
		return err
	}

	context := session.context(InboundDirection)
	if context.Disconnected {
		return ErrDisposed
	}

	return session.inbound.Handle(context, packet)
}

// Send handles and writes an outbound packet.
func (session *Session) Send(ctx context.Context, packet codec.Packet) error {
	if err := session.markTraffic(""); err != nil {
		return err
	}

	context := session.context(OutboundDirection)
	if context.Disconnected {
		return ErrDisposed
	}

	if err := session.outbound.Handle(context, packet); err != nil {
		return err
	}

	return session.sender(ctx, packet)
}

// Disconnect disposes the connection with a reason.
func (session *Session) Disconnect(ctx context.Context, reason Reason) error {
	session.mutex.Lock()
	if session.disconnected {
		session.mutex.Unlock()
		return ErrDisposed
	}

	session.disconnected = true
	session.disconnectReason = reason
	session.state = StateClosing
	close(session.done)
	disposer := session.disposer
	security := session.security
	session.mutex.Unlock()

	if security != nil {
		_ = security.Close(reason)
	}

	err := disposer(ctx, reason)

	session.mutex.Lock()
	session.state = StateClosed
	session.mutex.Unlock()

	return err
}

// Done returns a channel closed when the connection is disposed.
func (session *Session) Done() <-chan struct{} {
	return session.done
}
