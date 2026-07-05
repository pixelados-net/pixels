package connection

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/niflaot/pixels/internal/auth/sso"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inmachine "github.com/niflaot/pixels/networking/inbound/security/machine"
	inticket "github.com/niflaot/pixels/networking/inbound/security/ticket"
	outauth "github.com/niflaot/pixels/networking/outbound/authentication/ok"
	outping "github.com/niflaot/pixels/networking/outbound/client/ping"
	outidentity "github.com/niflaot/pixels/networking/outbound/handshake/identity"
	outmachine "github.com/niflaot/pixels/networking/outbound/security/machine"
	outstatus "github.com/niflaot/pixels/networking/outbound/session/hotel/availability/status"
)

// machineHandler handles machine identity packets.
func machineHandler(context netconn.Context, packet codec.Packet) error {
	payload, err := inmachine.Decode(packet)
	if err != nil {
		return err
	}

	if validMachine(payload.MachineID) {
		return nil
	}

	machineID, err := randomMachineID()
	if err != nil {
		return err
	}

	response, err := outmachine.Encode(machineID)
	if err != nil {
		return err
	}

	return context.Send(background(), response)
}

// ticketHandler handles SSO authentication packets.
func ticketHandler(service *sso.Service) netconn.Handler {
	return func(context netconn.Context, packet codec.Packet) error {
		payload, err := inticket.Decode(packet)
		if err != nil {
			return err
		}

		return authenticate(context, service, payload.Ticket)
	}
}

// authenticate consumes SSO and sends the initial bootstrap.
func authenticate(handler netconn.Context, service *sso.Service, ticket string) error {
	ctx := background()
	if err := handler.ValidateAuthenticationSecurity(ctx); err != nil {
		return err
	}

	if err := handler.Transition(netconn.EventAuthenticationStarted); err != nil {
		return err
	}

	if _, err := service.Consume(ctx, sso.ConsumeRequest{Ticket: ticket, IP: handler.RemoteAddr}); err != nil {
		_ = handler.Transition(netconn.EventAuthenticationRejected)
		return handler.Disconnect(ctx, netconn.Reason{Code: netconn.DisconnectAuthenticationFailed, Message: err.Error()})
	}

	if err := handler.Authenticate(time.Now()); err != nil {
		return err
	}

	return sendBootstrap(handler)
}

// sendBootstrap sends the minimal connection bootstrap.
func sendBootstrap(handler netconn.Context) error {
	for _, packet := range bootstrapPackets() {
		if err := handler.Send(background(), packet); err != nil {
			return err
		}
	}

	return handler.Transition(netconn.EventSessionReady)
}

// bootstrapPackets returns the first authenticated packets.
func bootstrapPackets() []codec.Packet {
	auth, _ := outauth.Encode()
	identity, _ := outidentity.Encode(0)
	status, _ := outstatus.Encode(true, false, outstatus.WithIsAuthentic(true))
	ping, _ := outping.Encode()

	return []codec.Packet{auth, identity, status, ping}
}

// validMachine reports whether a machine id is acceptable.
func validMachine(machineID string) bool {
	return len(machineID) == 64 && machineID[0] != '~'
}

// randomMachineID creates a hex machine id.
func randomMachineID() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return hex.EncodeToString(value), nil
}
