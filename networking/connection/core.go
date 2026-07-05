// Package connection defines transport-agnostic protocol sessions.
package connection

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/niflaot/pixels/networking/codec"
)

var (
	// ErrConnectionExists reports a duplicate connection id.
	ErrConnectionExists = errors.New("connection exists")
	// ErrConnectionNotFound reports a missing connection id.
	ErrConnectionNotFound = errors.New("connection not found")
	// ErrDisposed reports an operation after disposal.
	ErrDisposed = errors.New("connection disposed")
	// ErrHandlerExists reports a duplicate packet handler.
	ErrHandlerExists = errors.New("handler exists")
	// ErrHandlerNotFound reports a missing packet handler.
	ErrHandlerNotFound = errors.New("handler not found")
	// ErrHandlerPolicy reports a packet rejected by policy.
	ErrHandlerPolicy = errors.New("handler policy rejected")
	// ErrInvalidConnection reports an invalid connection.
	ErrInvalidConnection = errors.New("invalid connection")
	// ErrInvalidConnectionConfig reports invalid session config.
	ErrInvalidConnectionConfig = errors.New("invalid connection config")
	// ErrInvalidHandler reports an invalid packet handler.
	ErrInvalidHandler = errors.New("invalid handler")
	// ErrInvalidSecurity reports an invalid secure channel.
	ErrInvalidSecurity = errors.New("invalid security")
	// ErrInvalidState reports an invalid state operation.
	ErrInvalidState = errors.New("invalid state")
	// ErrInvalidTransition reports an invalid state transition.
	ErrInvalidTransition = errors.New("invalid transition")
	// ErrSecurityRequired reports missing required security.
	ErrSecurityRequired = errors.New("security required")
)

// ID identifies one connection within a connection kind.
type ID string

// Kind classifies the transport or session family of a connection.
type Kind string

// Sender writes an outbound packet through a transport.
type Sender func(context.Context, codec.Packet) error

// Disposer releases transport resources for a connection.
type Disposer func(context.Context, Reason) error

// PacketLogger records packet traffic and routing misses.
type PacketLogger interface {
	// Received records an inbound packet.
	Received(Context, codec.Packet)
	// Sent records an outbound packet.
	Sent(Context, codec.Packet)
	// Unhandled records a packet without a registered handler.
	Unhandled(Context, codec.Packet)
}

// SecurityActivator activates security after queued transport writes.
type SecurityActivator func(context.Context, SecureChannel) error

// Direction names whether a packet is entering or leaving a connection.
type Direction uint8

const (
	// InboundDirection describes a packet received from the peer.
	InboundDirection Direction = iota + 1

	// OutboundDirection describes a packet sent to the peer.
	OutboundDirection
)

// SecurityState names the byte security phase.
type SecurityState uint8

const (
	// SecurityPlain means traffic is not encrypted.
	SecurityPlain SecurityState = iota + 1

	// SecurityNegotiating means security is being negotiated.
	SecurityNegotiating

	// SecurityReady means encryption can open and seal bytes.
	SecurityReady

	// SecurityFailed means security negotiation failed.
	SecurityFailed
)

// SecurityMode names whether encryption is required.
type SecurityMode uint8

const (
	// SecurityOptional allows plain traffic before authentication.
	SecurityOptional SecurityMode = iota + 1

	// SecurityRequired requires secure traffic before authentication.
	SecurityRequired
)

// SecurityPolicy controls connection security requirements.
type SecurityPolicy struct {
	// Mode names whether security is optional or required.
	Mode SecurityMode
}

// DisconnectCode names a protocol-agnostic disconnection reason.
type DisconnectCode uint16

const (
	// DisconnectUnknown is used when no better reason is known.
	DisconnectUnknown DisconnectCode = iota
	// DisconnectLocalClose is used when the server closes intentionally.
	DisconnectLocalClose
	// DisconnectRemoteClose is used when the peer closes intentionally.
	DisconnectRemoteClose
	// DisconnectTransportError is used when the transport fails.
	DisconnectTransportError
	// DisconnectProtocolError is used when framing or payloads are invalid.
	DisconnectProtocolError
	// DisconnectAuthenticationFailed is used when authentication is rejected.
	DisconnectAuthenticationFailed
	// DisconnectAuthenticationTimeout is used when authentication takes too long.
	DisconnectAuthenticationTimeout
	// DisconnectDuplicateSession is used when a newer session replaces this one.
	DisconnectDuplicateSession
	// DisconnectIdleTimeout is used when the connection is idle too long.
	DisconnectIdleTimeout
	// DisconnectRateLimited is used when the peer exceeds network limits.
	DisconnectRateLimited
	// DisconnectPolicyViolation is used when the peer violates policy.
	DisconnectPolicyViolation
	// DisconnectKicked is used when moderation removes the peer.
	DisconnectKicked
	// DisconnectBanned is used when moderation blocks access.
	DisconnectBanned
	// DisconnectServerShutdown is used during controlled shutdown.
	DisconnectServerShutdown
)

// Reason explains why a connection was disconnected.
type Reason struct {
	// Code is the stable disconnection category.
	Code DisconnectCode
	// Message adds optional operator context.
	Message string
}

// SecureChannel opens and seals transport bytes for a session.
type SecureChannel interface {
	// State returns the security phase.
	State() SecurityState
	// Begin starts security negotiation.
	Begin(context.Context) error
	// Open unwraps inbound bytes.
	Open([]byte) ([]byte, error)
	// Seal wraps outbound bytes.
	Seal([]byte) ([]byte, error)
	// Close releases security state.
	Close(Reason) error
}

// DefaultSecurityPolicy returns the development-friendly policy.
func DefaultSecurityPolicy() SecurityPolicy {
	return SecurityPolicy{Mode: SecurityOptional}
}

// SecurityPolicyForEnvironment returns a policy for an environment name.
func SecurityPolicyForEnvironment(environment string) SecurityPolicy {
	if strings.EqualFold(environment, "production") {
		return SecurityPolicy{Mode: SecurityRequired}
	}

	return DefaultSecurityPolicy()
}

// Connection describes one transport-agnostic session.
type Connection interface {
	// ID returns the connection identifier.
	ID() ID
	// Kind returns the connection kind.
	Kind() Kind
	// StartedAt returns the connection start time.
	StartedAt() time.Time
	// AuthenticatedAt returns the authentication time when available.
	AuthenticatedAt() (time.Time, bool)
	// Authenticate marks the connection as authenticated.
	Authenticate(time.Time) error
	// State returns the connection lifecycle phase.
	State() State
	// Receive handles an inbound packet.
	Receive(context.Context, codec.Packet) error
	// Send handles and writes an outbound packet.
	Send(context.Context, codec.Packet) error
	// Disconnect disposes the connection with a reason.
	Disconnect(context.Context, Reason) error
	// Done returns a channel closed when the connection is disposed.
	Done() <-chan struct{}
}

// SessionConfig configures a session connection.
type SessionConfig struct {
	// ID identifies one connection within its kind.
	ID ID
	// Kind classifies the connection transport family.
	Kind Kind
	// StartedAt overrides the connection start time.
	StartedAt time.Time
	// RemoteAddr stores the transport peer address.
	RemoteAddr string
	// Inbound handles packets received from the peer.
	Inbound *HandlerRegistry
	// Outbound handles packets sent to the peer.
	Outbound *HandlerRegistry
	// SecurityPolicy controls whether encryption is required.
	SecurityPolicy SecurityPolicy
	// PacketLogger records development packet traffic.
	PacketLogger PacketLogger
	// Sender writes outbound packets through the transport.
	Sender Sender
	// Disposer releases transport resources.
	Disposer Disposer
	// SecurityActivator activates security after queued writes.
	SecurityActivator SecurityActivator
}
