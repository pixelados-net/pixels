package inventory

import (
	"context"

	currencybroadcast "github.com/niflaot/pixels/internal/realm/inventory/currency/broadcast"
	currencychanged "github.com/niflaot/pixels/internal/realm/inventory/currency/events/changed"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// RegisterCurrencyBroadcaster subscribes live currency packet projection.
func RegisterCurrencyBroadcaster(lifecycle fx.Lifecycle, subscriber bus.Subscriber, broadcaster *currencybroadcast.Broadcaster) error {
	subscription, err := subscriber.Subscribe(currencychanged.Name, bus.PriorityLow, broadcaster.Handle)
	if err != nil {
		return err
	}

	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		subscription.Unsubscribe()
		return nil
	}})

	return nil
}
