package runtime

import (
	"context"
	"time"

	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	netconn "github.com/niflaot/pixels/networking/connection"
	buildersstatus "github.com/niflaot/pixels/networking/outbound/subscription/builders/status"
	calendardata "github.com/niflaot/pixels/networking/outbound/subscription/calendar/data"
	giftnotification "github.com/niflaot/pixels/networking/outbound/subscription/gift/notification"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// RegisterBootstrap subscribes subscription state delivery after player connection.
func RegisterBootstrap(lifecycle fx.Lifecycle, subscriber bus.Subscriber, service *core.Service, players *playerlive.Registry, connections *netconn.Registry) error {
	subscription, err := subscriber.Subscribe(playerconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerconnected.Payload)
		if !ok || payload.PlayerID <= 0 {
			return nil
		}
		player, found := players.Find(payload.PlayerID)
		if !found {
			return nil
		}
		peer := player.Peer()
		connection, found := connections.Get(peer.ConnectionKind(), peer.ConnectionID())
		if !found {
			return nil
		}
		return sendBootstrap(ctx, service, connection, payload.PlayerID)
	})
	if err != nil {
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		subscription.Unsubscribe()
		return nil
	}})
	return nil
}

// sendBootstrap sends subscription compatibility and active reward state.
func sendBootstrap(ctx context.Context, service *core.Service, connection netconn.Connection, playerID int64) error {
	packet, err := buildersstatus.Encode()
	if err != nil {
		return err
	}
	if err := connection.Send(ctx, packet); err != nil {
		return err
	}
	membership, found, err := service.Membership(ctx, playerID)
	if err != nil {
		return err
	}
	if found && core.RemainingClubGifts(membership) > 0 {
		packet, err = giftnotification.Encode(core.RemainingClubGifts(membership))
		if err != nil {
			return err
		}
		if err := connection.Send(ctx, packet); err != nil {
			return err
		}
	}
	campaign, opened, found, err := service.ActiveCalendarData(ctx, playerID)
	if err != nil || !found {
		return err
	}
	current := int32(time.Since(campaign.StartDate) / (24 * time.Hour))
	packet, err = calendardata.Encode(campaign.Name, campaign.Image, current, campaign.DayCount, opened, bootstrapMissed(current, opened))
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// bootstrapMissed returns past unclaimed calendar days.
func bootstrapMissed(current int32, opened []int32) []int32 {
	claimed := make(map[int32]struct{}, len(opened))
	for _, day := range opened {
		claimed[day] = struct{}{}
	}
	missed := make([]int32, 0, current)
	for day := int32(0); day < current; day++ {
		if _, found := claimed[day]; !found {
			missed = append(missed, day)
		}
	}
	return missed
}
