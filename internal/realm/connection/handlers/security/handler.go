// Package security contains connection security and authentication handlers.
package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	currencyrequest "github.com/niflaot/pixels/internal/realm/inventory/currency/commands/request"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inmachine "github.com/niflaot/pixels/networking/inbound/security/machine"
	inticket "github.com/niflaot/pixels/networking/inbound/security/ticket"
	outauth "github.com/niflaot/pixels/networking/outbound/authentication/ok"
	outping "github.com/niflaot/pixels/networking/outbound/client/ping"
	outidentity "github.com/niflaot/pixels/networking/outbound/handshake/identity"
	outmachine "github.com/niflaot/pixels/networking/outbound/security/machine"
	outstatus "github.com/niflaot/pixels/networking/outbound/session/hotel/availability/status"
	outuserinfo "github.com/niflaot/pixels/networking/outbound/user/info"
)

// Register adds security handlers to a registry.
func Register(registry *netconn.HandlerRegistry, authenticator *Authenticator) {
	_ = registry.Register(inmachine.Header, Machine, netconn.AllowStates(netconn.StateHandshaking, netconn.StateSecuring), netconn.AllowUnauthenticated())
	_ = registry.Register(inticket.Header, Ticket(authenticator), netconn.AllowStates(netconn.StateHandshaking), netconn.AllowUnauthenticated())
}

// Machine handles machine identity packets.
func Machine(handler netconn.Context, packet codec.Packet) error {
	payload, err := inmachine.Decode(packet)
	if err != nil {
		return err
	}

	if validMachine(payload.MachineID) {
		return nil
	}

	return sendMachineReplacement(handler)
}

// Ticket handles SSO authentication packets.
func Ticket(authenticator *Authenticator) netconn.Handler {
	return func(handler netconn.Context, packet codec.Packet) error {
		payload, err := inticket.Decode(packet)
		if err != nil {
			return err
		}

		return authenticate(handler, authenticator, payload.Ticket)
	}
}

// authenticate consumes SSO and sends the initial bootstrap.
func authenticate(handler netconn.Context, authenticator *Authenticator, ticket string) error {
	ctx := context.Background()
	if err := handler.ValidateAuthenticationSecurity(ctx); err != nil {
		return err
	}

	if err := handler.Transition(netconn.EventAuthenticationStarted); err != nil {
		return err
	}

	record, err := authenticator.Resolve(ctx, handler, ticket)
	if err != nil {
		_ = handler.Transition(netconn.EventAuthenticationRejected)
		return handler.Disconnect(ctx, netconn.Reason{Code: netconn.DisconnectAuthenticationFailed, Message: err.Error()})
	}

	authenticatedAt := time.Now()
	if err := handler.Authenticate(authenticatedAt); err != nil {
		return err
	}

	if err := authenticator.Bind(ctx, handler, record, authenticatedAt); err != nil {
		_ = handler.Transition(netconn.EventAuthenticationRejected)
		return handler.Disconnect(ctx, netconn.Reason{Code: netconn.DisconnectAuthenticationFailed, Message: err.Error()})
	}

	if err := sendBootstrap(handler, record, authenticator); err != nil {
		return err
	}
	authenticator.Connected(ctx, handler, record.Player.ID)
	return nil
}

// sendBootstrap sends the minimal connection bootstrap.
func sendBootstrap(handler netconn.Context, record playerservice.Record, authenticator *Authenticator) error {
	for _, packet := range bootstrapPackets(record) {
		if err := handler.Send(context.Background(), packet); err != nil {
			return err
		}
	}
	if authenticator.currencies != nil {
		player, found := authenticator.live.Find(record.Player.ID)
		if !found {
			return currencyrequest.ErrPlayerNotFound
		}
		if err := authenticator.currencies.Send(context.Background(), handler, player.Currencies()); err != nil {
			return err
		}
	}
	if authenticator.permissions != nil {
		packets, err := authenticator.permissions.Packets(context.Background(), record.Player.ID)
		if err != nil {
			return err
		}
		for _, packet := range packets {
			if err := handler.Send(context.Background(), packet); err != nil {
				return err
			}
		}
	}

	return handler.Transition(netconn.EventSessionReady)
}

// sendMachineReplacement sends a generated machine identifier.
func sendMachineReplacement(handler netconn.Context) error {
	machineID, err := randomMachineID()
	if err != nil {
		return err
	}

	response, err := outmachine.Encode(machineID)
	if err != nil {
		return err
	}

	return handler.Send(context.Background(), response)
}

// bootstrapPackets returns the first authenticated packets.
func bootstrapPackets(record playerservice.Record) []codec.Packet {
	auth, _ := outauth.Encode()
	userInfo, _ := outuserinfo.Encode(outuserinfo.Params{
		UserID:        int32(record.Player.ID),
		Username:      record.Player.Username,
		Figure:        record.Profile.Look,
		Gender:        string(record.Profile.Gender),
		Motto:         record.Profile.Motto,
		CanChangeName: record.Profile.AllowNameChange,
	})
	identity, _ := outidentity.Encode(0)
	status, _ := outstatus.Encode(true, false, outstatus.WithIsAuthentic(true))
	ping, _ := outping.Encode()

	return []codec.Packet{auth, userInfo, identity, status, ping}
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
