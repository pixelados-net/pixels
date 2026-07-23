package trade

import (
	"context"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	roomleft "github.com/niflaot/pixels/internal/realm/room/access/events/left"
	tradeadmin "github.com/niflaot/pixels/internal/realm/trade/admin"
	tradeconfirm "github.com/niflaot/pixels/internal/realm/trade/confirm"
	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	tradedb "github.com/niflaot/pixels/internal/realm/trade/database"
	tradeoffer "github.com/niflaot/pixels/internal/realm/trade/offer"
	traderecord "github.com/niflaot/pixels/internal/realm/trade/record"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	tradesession "github.com/niflaot/pixels/internal/realm/trade/session"
	outclosed "github.com/niflaot/pixels/networking/outbound/trade/closed"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// Module provides direct-trade persistence, runtime, and packet behavior.
var Module = fx.Module("realm-trade", fx.Provide(LoadConfig, tradedb.New, NewStore, traderuntime.NewRegistry, traderuntime.NewSender, tradecore.New, tradeadmin.New), fx.Invoke(ConfigureFurniture, RegisterHandlers, RegisterDisconnects))

// NewStore exposes trade persistence behavior.
func NewStore(repository *tradedb.Repository) traderecord.Store { return repository }

// ConfigureFurniture attaches live staged-item checks to furniture behavior.
func ConfigureFurniture(furniture *furnitureservice.Service, registry *traderuntime.Registry) {
	furniture.SetStagedChecker(registry)
}

// RegisterHandlers installs all direct-trade packet adapters.
func RegisterHandlers(handlers *realmconn.Handlers, service *tradecore.Service, sender *traderuntime.Sender, furniture furnitureservice.TradingManager) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	tradesession.Register(handlers.Inbound, tradesession.Handler{Service: service, Sender: sender})
	tradeoffer.Register(handlers.Inbound, tradeoffer.Handler{Service: service, Sender: sender, Furniture: furniture})
	tradeconfirm.Register(handlers.Inbound, tradeconfirm.Handler{Service: service, Sender: sender})
}

// RegisterDisconnects cancels direct trades when either player disconnects.
func RegisterDisconnects(lifecycle fx.Lifecycle, subscriber bus.Subscriber, service *tradecore.Service, sender *traderuntime.Sender) error {
	disconnectedSubscription, err := subscriber.Subscribe(playerdisconnected.Name, bus.PriorityHigh, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerdisconnected.Payload)
		if ok {
			cancelForLifecycle(ctx, service, sender, payload.PlayerID)
		}
		return nil
	})
	if err != nil {
		return err
	}
	leftSubscription, err := subscriber.Subscribe(roomleft.Name, bus.PriorityHigh, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(roomleft.Payload)
		if ok {
			cancelForLifecycle(ctx, service, sender, payload.PlayerID)
		}
		return nil
	})
	if err != nil {
		disconnectedSubscription.Unsubscribe()
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		disconnectedSubscription.Unsubscribe()
		leftSubscription.Unsubscribe()
		return nil
	}})
	return nil
}

// cancelForLifecycle closes a stale room trade and notifies every reachable participant.
func cancelForLifecycle(ctx context.Context, service *tradecore.Service, sender *traderuntime.Sender, playerID int64) {
	session, found := service.Registry().Find(playerID)
	if !found {
		return
	}
	participant, _ := session.Participant(playerID)
	if !service.Close(playerID) {
		return
	}
	packet, _ := outclosed.Encode(participant.PlayerID, 0)
	_ = sender.Send(ctx, session.First.PlayerID, packet)
	_ = sender.Send(ctx, session.Second.PlayerID, packet)
}
