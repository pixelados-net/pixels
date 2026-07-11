package room

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/leave"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	moderationbroadcast "github.com/niflaot/pixels/internal/realm/room/control/moderation/broadcast"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomoccupancy "github.com/niflaot/pixels/internal/realm/room/runtime/events/occupancychanged"
	"github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
)

// NewLiveRegistry creates the active room registry.
func NewLiveRegistry(publisher bus.Publisher, connections *netconn.Registry, players *playerlive.Registry, config roomentry.Config, entryService *roomentry.Service, translations i18n.Translator) *live.Registry {
	var registry *live.Registry
	registry = live.NewRegistry(func(ctx context.Context, occupancy live.Occupancy) error {
		return publisher.Publish(ctx, bus.Event{Name: roomoccupancy.Name, Payload: occupancyEvent(occupancy)})
	},
		live.WithMovementPublisher(newMovementPublisher(connections, players, publisher, translations, func() *live.Registry { return registry })),
		live.WithDoorbellPublisher(broadcast.NewDoorbellPublisher(connections)),
		live.WithDoorbellApprover(newDoorbellApprover(entryService)),
		live.WithDoorbellTimeout(config.Normalize().HangoutTimeout),
	)

	return registry
}

// newMovementPublisher broadcasts movement before completing door exits.
func newMovementPublisher(connections *netconn.Registry, players *playerlive.Registry, publisher bus.Publisher, translations i18n.Translator, registry func() *live.Registry) live.MovementPublisher {
	broadcastMovement := broadcast.NewMovementPublisher(connections)

	return func(ctx context.Context, active *live.Room, movements []live.Movement) error {
		movementErr := broadcastMovement(ctx, active, movements)
		leave := leavecmd.Handler{Players: players, Runtime: registry(), Connections: connections, Events: publisher}
		for _, movement := range movements {
			if !movement.Exited {
				continue
			}
			if movement.ForcedExit {
				notice, err := moderationbroadcast.KickedNotice(translations)
				if err != nil {
					movementErr = errors.Join(movementErr, leave.ToDesktop(ctx, movement.PlayerID), err)
					continue
				}
				movementErr = errors.Join(movementErr, leave.ToDesktopThen(ctx, movement.PlayerID, notice))
			} else {
				movementErr = errors.Join(movementErr, leave.ToDesktop(ctx, movement.PlayerID))
			}
		}

		return movementErr
	}
}

// newDoorbellApprover creates a room responder presence check.
func newDoorbellApprover(entryService *roomentry.Service) live.DoorbellApprover {
	if entryService == nil {
		return func(_ context.Context, active *live.Room) (bool, error) {
			return active.OwnerPresent(), nil
		}
	}

	return func(ctx context.Context, active *live.Room) (bool, error) {
		snapshot := active.Snapshot()
		for _, occupant := range active.Occupants() {
			allowed, err := entryService.CanAnswerDoorbell(ctx, snapshot.ID, snapshot.OwnerPlayerID, occupant.PlayerID)
			if err != nil || allowed {
				return allowed, err
			}
		}

		return false, nil
	}
}

// RegisterRuntimeCleanup removes room occupancy on player disconnect.
func RegisterRuntimeCleanup(lifecycle fx.Lifecycle, subscriber bus.Subscriber, publisher bus.Publisher, registry *live.Registry, connections *netconn.Registry) error {
	subscription, err := subscriber.Subscribe(playerdisconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		disconnected, ok := event.Payload.(playerdisconnected.Payload)
		if !ok || disconnected.PlayerID <= 0 {
			return nil
		}

		return (leavecmd.Handler{
			Runtime: registry, Connections: connections, Events: publisher,
		}).Handle(ctx, command.Envelope[leavecmd.Command]{
			Command: leavecmd.Command{PlayerID: disconnected.PlayerID},
		})
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

// occupancyEvent maps live occupancy to a realm event payload.
func occupancyEvent(occupancy live.Occupancy) roomoccupancy.Payload {
	return roomoccupancy.Payload{
		RoomID:     occupancy.RoomID,
		CategoryID: occupancy.CategoryID,
		Count:      occupancy.Count,
		MaxUsers:   occupancy.MaxUsers,
		PlayerIDs:  append([]int64(nil), occupancy.PlayerIDs...),
	}
}
