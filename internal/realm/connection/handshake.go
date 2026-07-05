package connection

import (
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

// releaseHandler handles client release metadata.
func releaseHandler(context netconn.Context, packet codec.Packet) error {
	_, err := inrelease.Decode(packet)

	return err
}

// variablesHandler handles client variable metadata.
func variablesHandler(context netconn.Context, packet codec.Packet) error {
	_, err := invariables.Decode(packet)

	return err
}

// policyHandler handles client policy probes.
func policyHandler(context netconn.Context, packet codec.Packet) error {
	_, err := inpolicy.Decode(packet)

	return err
}

// diffieInitHandler handles Diffie start packets.
func diffieInitHandler(context netconn.Context, packet codec.Packet) error {
	if _, err := indiffieinit.Decode(packet); err != nil {
		return err
	}

	_ = context.Transition(netconn.EventProtocolFailed)

	return context.Disconnect(background(), netconn.Reason{Code: netconn.DisconnectProtocolError, Message: ErrDiffieUnavailable.Error()})
}

// diffieCompleteHandler handles Diffie completion packets.
func diffieCompleteHandler(context netconn.Context, packet codec.Packet) error {
	if _, err := indiffiecomplete.Decode(packet); err != nil {
		return err
	}

	_ = context.Transition(netconn.EventProtocolFailed)

	return context.Disconnect(background(), netconn.Reason{Code: netconn.DisconnectProtocolError, Message: ErrDiffieUnavailable.Error()})
}
