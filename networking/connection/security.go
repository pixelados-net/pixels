package connection

import (
	"context"
	"time"

	"github.com/niflaot/pixels/networking/codec"
)

// Send writes an outbound packet through the current session.
func (context Context) Send(ctx context.Context, packet codec.Packet) error {
	if context.session == nil {
		return ErrInvalidConnection
	}

	return context.session.Send(ctx, packet)
}

// Disconnect disposes the current session with a reason.
func (context Context) Disconnect(ctx context.Context, reason Reason) error {
	if context.session == nil {
		return ErrInvalidConnection
	}

	return context.session.Disconnect(ctx, reason)
}

// Authenticate marks the current session authenticated.
func (context Context) Authenticate(at time.Time) error {
	if context.session == nil {
		return ErrInvalidConnection
	}

	return context.session.Authenticate(at)
}

// Transition applies a lifecycle event to the current session.
func (context Context) Transition(event Event) error {
	if context.session == nil {
		return ErrInvalidConnection
	}

	return context.session.Transition(event)
}

// MarkPong records a heartbeat pong for the current session.
func (context Context) MarkPong(at time.Time) error {
	if context.session == nil {
		return ErrInvalidConnection
	}

	return context.session.MarkPong(at)
}

// ValidateAuthenticationSecurity checks security before authentication.
func (context Context) ValidateAuthenticationSecurity(ctx context.Context) error {
	if context.session == nil {
		return ErrInvalidConnection
	}

	return context.session.ValidateAuthenticationSecurity(ctx)
}

// CompleteSecurity sends completion and activates security after queued writes.
func (context Context) CompleteSecurity(ctx context.Context, packet codec.Packet, channel SecureChannel) error {
	if context.session == nil {
		return ErrInvalidConnection
	}

	return context.session.CompleteSecurity(ctx, packet, channel)
}

// SecurityPolicy returns the connection security policy.
func (session *Session) SecurityPolicy() SecurityPolicy {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.policy
}

// SetSecurityPolicy changes security policy before traffic starts.
func (session *Session) SetSecurityPolicy(policy SecurityPolicy) error {
	session.mutex.Lock()
	defer session.mutex.Unlock()

	if session.trafficStarted {
		return ErrInvalidState
	}

	session.policy = normalizeSecurityPolicy(policy)

	return nil
}

// AttachSecurity attaches a secure channel to the session.
func (session *Session) AttachSecurity(channel SecureChannel) error {
	if channel == nil {
		return ErrInvalidSecurity
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	if session.disconnected {
		return ErrDisposed
	}

	if session.security != nil {
		return ErrInvalidSecurity
	}

	session.security = channel

	return nil
}

// CompleteSecurity sends a plaintext completion packet before activating security.
func (session *Session) CompleteSecurity(ctx context.Context, packet codec.Packet, channel SecureChannel) error {
	if channel == nil {
		return ErrInvalidSecurity
	}

	if err := session.Send(ctx, packet); err != nil {
		return err
	}

	activator := session.securityActivator()
	if activator == nil {
		return session.AttachSecurity(channel)
	}

	return activator(ctx, channel)
}

// SecurityState returns the attached secure channel state.
func (session *Session) SecurityState() SecurityState {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	if session.security == nil {
		return SecurityPlain
	}

	return session.security.State()
}

// Open unwraps inbound bytes when security is ready.
func (session *Session) Open(src []byte) ([]byte, error) {
	channel := session.secureChannel()
	if channel == nil || channel.State() != SecurityReady {
		return src, nil
	}

	return channel.Open(src)
}

// Seal wraps outbound bytes when security is ready.
func (session *Session) Seal(src []byte) ([]byte, error) {
	channel := session.secureChannel()
	if channel == nil || channel.State() != SecurityReady {
		return src, nil
	}

	return channel.Seal(src)
}

// ValidateAuthenticationSecurity checks security before authentication.
func (session *Session) ValidateAuthenticationSecurity(ctx context.Context) error {
	if session.SecurityPolicy().Mode != SecurityRequired {
		return nil
	}

	if session.SecurityState() == SecurityReady {
		return nil
	}

	_ = session.Transition(EventProtocolFailed)
	_ = session.Disconnect(ctx, Reason{Code: DisconnectProtocolError, Message: ErrSecurityRequired.Error()})

	return ErrSecurityRequired
}

// ActivateSecurity attaches a secure channel from a transport barrier.
func (session *Session) ActivateSecurity(channel SecureChannel) error {
	return session.AttachSecurity(channel)
}

// secureChannel returns the attached secure channel.
func (session *Session) secureChannel() SecureChannel {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.security
}

// securityActivator returns the transport activation hook.
func (session *Session) securityActivator() SecurityActivator {
	session.mutex.RLock()
	defer session.mutex.RUnlock()

	return session.activator
}
