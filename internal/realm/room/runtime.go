package room

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/leave"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	moderationbroadcast "github.com/niflaot/pixels/internal/realm/room/control/moderation/broadcast"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomoccupancy "github.com/niflaot/pixels/internal/realm/room/runtime/events/occupancychanged"
	"github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roommoved "github.com/niflaot/pixels/internal/realm/room/world/events/moved"
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
			if movement.Moved {
				movementErr = errors.Join(movementErr, publishFurnitureSteps(ctx, publisher, active, movement))
				movementErr = errors.Join(movementErr, publishUnitMoved(ctx, publisher, active, movement))
			}
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

// publishUnitMoved emits one post-commit movement event for room capabilities.
func publishUnitMoved(ctx context.Context, publisher bus.Publisher, active *live.Room, movement live.Movement) error {
	if publisher == nil {
		return nil
	}
	return publisher.Publish(ctx, bus.Event{Name: roommoved.Name, Payload: roommoved.Payload{
		RoomID: active.ID(), EntityKey: movement.Unit.EntityKey, PlayerID: movement.Unit.PlayerID,
		Kind: movement.Unit.Kind, Previous: movement.Unit.Previous, Current: movement.Unit.Position,
	}})
}

// publishFurnitureSteps emits walk-off and walk-on events from accepted movement.
func publishFurnitureSteps(ctx context.Context, publisher bus.Publisher, active *live.Room, movement live.Movement) error {
	if publisher == nil {
		return nil
	}
	var previousStorage [16]int64
	var currentStorage [16]int64
	previous := active.FurnitureIDsAt(movement.Unit.Previous.Point, previousStorage[:0])
	current := active.FurnitureIDsAt(movement.Unit.Position.Point, currentStorage[:0])
	var result error
	for _, itemID := range previous {
		if containsItemID(current, itemID) {
			continue
		}
		result = publisher.Publish(ctx, bus.Event{Name: furniturewalkedoff.Name, Payload: furniturewalkedoff.Payload{
			PlayerID: movement.PlayerID, ItemID: itemID, RoomID: active.ID(),
		}})
	}
	for _, itemID := range current {
		if containsItemID(previous, itemID) {
			continue
		}
		result = errors.Join(result, publisher.Publish(ctx, bus.Event{Name: furniturewalkedon.Name, Payload: furniturewalkedon.Payload{
			PlayerID: movement.PlayerID, ItemID: itemID, RoomID: active.ID(),
		}}))
	}

	return result
}

// containsItemID reports whether an item appears in one tile snapshot.
func containsItemID(items []int64, itemID int64) bool {
	for _, current := range items {
		if current == itemID {
			return true
		}
	}
	return false
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
