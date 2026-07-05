package security

import (
	"context"
	"fmt"
	"time"

	"github.com/niflaot/pixels/internal/auth/sso"
	playerrealm "github.com/niflaot/pixels/internal/realm/player"
	"github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	sessionrealm "github.com/niflaot/pixels/internal/realm/session"
	"github.com/niflaot/pixels/internal/realm/session/binding"
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
}

// NewAuthenticator creates a security authenticator.
func NewAuthenticator(tickets *sso.Service, players playerservice.Finder, live *live.Registry, bindings *binding.Registry, events bus.Publisher) *Authenticator {
	return &Authenticator{
		tickets:  tickets,
		players:  players,
		live:     live,
		bindings: bindings,
		events:   events,
	}
}

// Resolve consumes a ticket and loads the bound player record.
func (authenticator *Authenticator) Resolve(ctx context.Context, handler netconn.Context, ticketValue string) (playerservice.Record, error) {
	ticket, err := authenticator.tickets.Consume(ctx, sso.ConsumeRequest{Ticket: ticketValue, IP: handler.RemoteAddr})
	if err != nil {
		authenticator.publish(ctx, playerrealm.EventAuthenticationFailed, authenticationEvent(handler, 0, err.Error()))
		return playerservice.Record{}, err
	}

	authenticator.publish(ctx, playerrealm.EventAuthenticating, authenticationEvent(handler, ticket.PlayerID, ""))

	record, found, err := authenticator.players.FindByID(ctx, ticket.PlayerID)
	if err != nil {
		authenticator.publish(ctx, playerrealm.EventAuthenticationFailed, authenticationEvent(handler, ticket.PlayerID, err.Error()))
		return playerservice.Record{}, fmt.Errorf("load player %d: %w", ticket.PlayerID, err)
	}
	if !found {
		authenticator.publish(ctx, playerrealm.EventAuthenticationFailed, authenticationEvent(handler, ticket.PlayerID, playerservice.ErrPlayerNotFound.Error()))
		return playerservice.Record{}, playerservice.ErrPlayerNotFound
	}

	authenticator.publish(ctx, playerrealm.EventProfileLoaded, authenticationEvent(handler, ticket.PlayerID, ""))

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

	authenticator.publish(ctx, sessionrealm.EventBound, sessionrealm.BindingEvent{Binding: sessionBinding})
	authenticator.publish(ctx, playerrealm.EventAuthenticated, authenticationEvent(handler, record.Player.ID, ""))
	authenticator.publish(ctx, playerrealm.EventConnected, authenticationEvent(handler, record.Player.ID, ""))

	return nil
}

// publish emits an event when an event bus is configured.
func (authenticator *Authenticator) publish(ctx context.Context, name bus.Name, payload any) {
	if authenticator.events == nil {
		return
	}

	_ = authenticator.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
}

// authenticationEvent creates a player authentication event payload.
func authenticationEvent(handler netconn.Context, playerID int64, reason string) playerrealm.AuthenticationEvent {
	return playerrealm.AuthenticationEvent{
		PlayerID:       playerID,
		ConnectionID:   handler.ConnectionID,
		ConnectionKind: handler.ConnectionKind,
		Reason:         reason,
	}
}
