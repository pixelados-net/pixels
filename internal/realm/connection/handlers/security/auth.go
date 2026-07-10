package security

import (
	"context"
	"fmt"
	"time"

	"github.com/niflaot/pixels/internal/auth/sso"
	permissionbroadcast "github.com/niflaot/pixels/internal/permission/broadcast"
	currencyrequest "github.com/niflaot/pixels/internal/realm/inventory/currency/commands/request"
	playerauthenticated "github.com/niflaot/pixels/internal/realm/player/events/authenticated"
	playerauthenticating "github.com/niflaot/pixels/internal/realm/player/events/authenticating"
	playerauthfailed "github.com/niflaot/pixels/internal/realm/player/events/authfailed"
	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	playerprofileloaded "github.com/niflaot/pixels/internal/realm/player/events/profileloaded"
	"github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	sessionbound "github.com/niflaot/pixels/internal/realm/session/events/bound"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// Authenticator resolves SSO tickets into live player sessions.
type Authenticator struct {
	// tickets consumes SSO tickets.
	tickets *sso.Service
	// players loads persistent player records.
	players playerservice.Finder
	// live stores online player runtime state.
	live *live.Registry
	// bindings stores player connection bindings.
	bindings *binding.Registry
	// events publishes authentication lifecycle events.
	events bus.Publisher
	// currencies sends the composed player's wallet bootstrap.
	currencies *currencyrequest.Handler
	// permissions sends the player's permission and perk bootstrap.
	permissions *permissionbroadcast.Projector
}

// NewAuthenticator creates a security authenticator.
func NewAuthenticator(tickets *sso.Service, players playerservice.Finder, live *live.Registry, bindings *binding.Registry, events bus.Publisher, currencies *currencyrequest.Handler, projectors ...*permissionbroadcast.Projector) *Authenticator {
	authenticator := &Authenticator{
		tickets:    tickets,
		players:    players,
		live:       live,
		bindings:   bindings,
		events:     events,
		currencies: currencies,
	}
	if len(projectors) > 0 {
		authenticator.permissions = projectors[0]
	}

	return authenticator
}

// Resolve consumes a ticket and loads the bound player record.
func (authenticator *Authenticator) Resolve(ctx context.Context, handler netconn.Context, ticketValue string) (playerservice.Record, error) {
	ticket, err := authenticator.tickets.Consume(ctx, sso.ConsumeRequest{Ticket: ticketValue, IP: handler.RemoteAddr})
	if err != nil {
		authenticator.publish(ctx, playerauthfailed.Name, playerauthfailed.Payload(authenticationPayloadFromHandler(handler, 0, err.Error())))
		return playerservice.Record{}, err
	}

	authenticator.publish(ctx, playerauthenticating.Name, playerauthenticating.Payload(authenticationPayloadFromHandler(handler, ticket.PlayerID, "")))

	record, found, err := authenticator.players.FindByID(ctx, ticket.PlayerID)
	if err != nil {
		authenticator.publish(ctx, playerauthfailed.Name, playerauthfailed.Payload(authenticationPayloadFromHandler(handler, ticket.PlayerID, err.Error())))
		return playerservice.Record{}, fmt.Errorf("load player %d: %w", ticket.PlayerID, err)
	}
	if !found {
		authenticator.publish(ctx, playerauthfailed.Name, playerauthfailed.Payload(authenticationPayloadFromHandler(handler, ticket.PlayerID, playerservice.ErrPlayerNotFound.Error())))
		return playerservice.Record{}, playerservice.ErrPlayerNotFound
	}

	authenticator.publish(ctx, playerprofileloaded.Name, playerprofileloaded.Payload(authenticationPayloadFromHandler(handler, ticket.PlayerID, "")))

	return record, nil
}

// Bind registers runtime player state for an authenticated connection.
func (authenticator *Authenticator) Bind(ctx context.Context, handler netconn.Context, record playerservice.Record, authenticatedAt time.Time) error {
	peer, err := live.NewSessionPeer(handler.ConnectionID, handler.ConnectionKind, authenticatedAt)
	if err != nil {
		return err
	}

	player, err := live.NewPlayer(live.SnapshotFromRecord(record), peer)
	if err != nil {
		return err
	}

	if err := authenticator.live.Add(player); err != nil {
		return err
	}

	sessionBinding := binding.Binding{
		PlayerID:       record.Player.ID,
		ConnectionID:   handler.ConnectionID,
		ConnectionKind: handler.ConnectionKind,
		BoundAt:        authenticatedAt,
	}
	if err := authenticator.bindings.Add(sessionBinding); err != nil {
		authenticator.live.Remove(record.Player.ID)
		return err
	}

	authenticator.publish(ctx, sessionbound.Name, sessionbound.Payload{Binding: sessionBinding})
	authenticator.publish(ctx, playerauthenticated.Name, playerauthenticated.Payload(authenticationPayloadFromHandler(handler, record.Player.ID, "")))
	authenticator.publish(ctx, playerconnected.Name, playerconnected.Payload(authenticationPayloadFromHandler(handler, record.Player.ID, "")))

	return nil
}

// publish emits an event when an event bus is configured.
func (authenticator *Authenticator) publish(ctx context.Context, name bus.Name, payload any) {
	if authenticator.events == nil {
		return
	}

	_ = authenticator.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
}

// authenticationPayload describes shared authentication event fields.
type authenticationPayload struct {
	// PlayerID identifies the player when known.
	PlayerID int64

	// ConnectionID identifies the connection.
	ConnectionID netconn.ID

	// ConnectionKind identifies the connection family.
	ConnectionKind netconn.Kind

	// Reason stores a failure reason when available.
	Reason string
}

// authenticationPayloadFromHandler creates shared authentication event fields.
func authenticationPayloadFromHandler(handler netconn.Context, playerID int64, reason string) authenticationPayload {
	return authenticationPayload{
		PlayerID:       playerID,
		ConnectionID:   handler.ConnectionID,
		ConnectionKind: handler.ConnectionKind,
		Reason:         reason,
	}
}
