// Package handshake contains connection handshake packet handlers.
package handshake

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	indiffiecomplete "github.com/niflaot/pixels/networking/inbound/handshake/diffie/complete"
	indiffieinit "github.com/niflaot/pixels/networking/inbound/handshake/diffie/init"
	inpolicy "github.com/niflaot/pixels/networking/inbound/handshake/policy"
	inrelease "github.com/niflaot/pixels/networking/inbound/handshake/release"
	invariables "github.com/niflaot/pixels/networking/inbound/handshake/variables"
)

var (
	// ErrDiffieUnavailable reports missing Diffie support.
	ErrDiffieUnavailable = errors.New("diffie unavailable")
)

// Register adds handshake handlers to a registry.
func Register(registry *netconn.HandlerRegistry) {
	early := []netconn.HandlerOption{netconn.AllowStates(netconn.StateCreated, netconn.StateHandshaking), netconn.AllowUnauthenticated()}
	_ = registry.Register(inrelease.Header, Release, early...)
	_ = registry.Register(invariables.Header, Variables, early...)
	_ = registry.Register(inpolicy.Header, Policy, early...)
	_ = registry.Register(indiffieinit.Header, DiffieInit, early...)
	_ = registry.Register(indiffiecomplete.Header, DiffieComplete, netconn.AllowStates(netconn.StateSecuring), netconn.AllowUnauthenticated())
}

// Release handles client release metadata.
func Release(context netconn.Context, packet codec.Packet) error {
	_, err := inrelease.Decode(packet)
	// TODO: Store this on connection.
	return err
}

// Variables handles client variable metadata.
func Variables(context netconn.Context, packet codec.Packet) error {
	_, err := invariables.Decode(packet)
	// TODO: Store this on connection.
	return err
}

// Policy handles client policy probes.
func Policy(context netconn.Context, packet codec.Packet) error {
	_, err := inpolicy.Decode(packet)
	// TODO: Ideate an implementaiton of policy handling.
	return err
}

// DiffieInit handles Diffie start packets.
func DiffieInit(context netconn.Context, packet codec.Packet) error {
	if _, err := indiffieinit.Decode(packet); err != nil {
		return err
	}

	return disconnectMissingDiffie(context)
}

// DiffieComplete handles Diffie completion packets.
func DiffieComplete(context netconn.Context, packet codec.Packet) error {
	if _, err := indiffiecomplete.Decode(packet); err != nil {
		return err
	}

	return disconnectMissingDiffie(context)
}

// disconnectMissingDiffie closes a session when Diffie is not implemented.
func disconnectMissingDiffie(handler netconn.Context) error {
	_ = handler.Transition(netconn.EventProtocolFailed)

	return handler.Disconnect(context.Background(), netconn.Reason{Code: netconn.DisconnectProtocolError, Message: ErrDiffieUnavailable.Error()})
}
