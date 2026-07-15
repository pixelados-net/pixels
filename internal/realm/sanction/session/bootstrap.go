// Package session owns sanction authentication and login projection.
package session

import (
	"context"
	"time"

	realmconnection "github.com/niflaot/pixels/internal/realm/connection"
	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	netconn "github.com/niflaot/pixels/networking/connection"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// ConfigureAuthentication installs active-ban validation before session binding.
func ConfigureAuthentication(handlers *realmconnection.Handlers, service *sanctioncore.Service) {
	handlers.SetSanctionGate(service)
}

// RegisterLifecycle hydrates live sanctions and drains pending warnings at login.
func RegisterLifecycle(lifecycle fx.Lifecycle, subscriber bus.Subscriber, service *sanctioncore.Service, players *playerlive.Registry, connections *netconn.Registry) error {
	subscription, err := subscriber.Subscribe(playerconnected.Name, bus.PriorityHigh, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerconnected.Payload)
		if !ok {
			return nil
		}
		return bootstrap(ctx, payload.PlayerID, service, players, connections)
	})
	if err != nil {
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { subscription.Unsubscribe(); return nil }})
	return nil
}

// bootstrap loads one sanction snapshot and queued warnings.
func bootstrap(ctx context.Context, playerID int64, service *sanctioncore.Service, players *playerlive.Registry, connections *netconn.Registry) error {
	state, err := service.Active(ctx, playerID)
	if err != nil {
		return err
	}
	player, found := players.Find(playerID)
	if !found {
		return nil
	}
	projection := playerlive.Sanctions{MutePermanent: state.MutedPermanently, TradeLockPermanent: state.TradeLockedPermanently}
	if state.MuteUntil != nil {
		projection.MuteUntil = *state.MuteUntil
	}
	if state.TradeLockUntil != nil {
		projection.TradeLockUntil = *state.TradeLockUntil
	}
	player.SetSanctions(projection)
	alerts, err := service.Store().PendingAlerts(ctx, playerID, 100)
	if err != nil {
		return err
	}
	peer := player.Peer()
	connection, found := connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if !found {
		return nil
	}
	for _, alert := range alerts {
		packet, encodeErr := outalert.Encode(alert.Message)
		if encodeErr != nil {
			continue
		}
		if sendErr := connection.Send(ctx, packet); sendErr != nil {
			continue
		}
		_ = service.Store().MarkAlertDelivered(ctx, alert.ID, time.Now())
	}
	return nil
}

// SanctionsAt maps an active state for focused tests.
func SanctionsAt(state sanctionrecord.ActiveState, now time.Time) playerlive.Sanctions {
	projection := playerlive.Sanctions{MutePermanent: state.MutedPermanently, TradeLockPermanent: state.TradeLockedPermanently}
	if state.MuteUntil != nil && state.MuteUntil.After(now) {
		projection.MuteUntil = *state.MuteUntil
	}
	if state.TradeLockUntil != nil && state.TradeLockUntil.After(now) {
		projection.TradeLockUntil = *state.TradeLockUntil
	}
	return projection
}
