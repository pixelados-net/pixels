package teleport

import (
	"context"
	"time"

	teleportfailed "github.com/niflaot/pixels/internal/realm/furniture/events/teleportfailed"
	teleportstarted "github.com/niflaot/pixels/internal/realm/furniture/events/teleportstarted"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

const pendingTTL = 30 * time.Second

// forward hands cross-room navigation to Nitro and records the destination spawn.
func (service *Service) forward(ctx context.Context, active *roomlive.Room, transit Transit) error {
	connection, found := service.playerConnection(active, transit.PlayerID)
	if !found {
		return service.fail(ctx, active.ID(), transit, "connection_not_found")
	}
	if service.config.BypassLocked && service.entry != nil {
		service.entry.GrantTrusted(transit.PlayerID, transit.TargetRoomID)
	}
	service.pendingMutex.Lock()
	if service.pending == nil {
		service.pending = make(map[int64]pendingDestination)
	}
	service.pending[transit.PlayerID] = pendingDestination{
		roomID: transit.TargetRoomID, transit: transit, expiresAt: service.now().Add(pendingTTL),
	}
	service.pendingMutex.Unlock()
	packet, err := outforward.Encode(int32(transit.TargetRoomID))
	if err != nil {
		return err
	}
	if err := connection.Send(ctx, packet); err != nil {
		service.clearPending(transit.PlayerID)
		return err
	}

	return nil
}

// playerConnection resolves one room occupant connection without a global player scan.
func (service *Service) playerConnection(active *roomlive.Room, playerID int64) (netconn.Connection, bool) {
	for _, occupant := range active.Occupants() {
		if occupant.PlayerID == playerID {
			return service.connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		}
	}

	return nil, false
}

// consumePending consumes a valid destination for one room entry.
func (service *Service) consumePending(playerID int64, roomID int64) (pendingDestination, bool) {
	service.pendingMutex.Lock()
	pending, found := service.pending[playerID]
	if !found {
		service.pendingMutex.Unlock()
		return pendingDestination{}, false
	}
	if pending.roomID != roomID {
		delete(service.pending, playerID)
		if len(service.pending) == 0 {
			service.pending = nil
		}
		service.pendingMutex.Unlock()
		service.releasePair(pending.transit)
		return pendingDestination{}, false
	}
	delete(service.pending, playerID)
	if len(service.pending) == 0 {
		service.pending = nil
	}
	service.pendingMutex.Unlock()
	if !pending.expiresAt.After(service.now()) {
		service.releasePair(pending.transit)
		return pendingDestination{}, false
	}

	return pending, true
}

// clearPending removes a player's pending destination.
func (service *Service) clearPending(playerID int64) {
	service.pendingMutex.Lock()
	pending, found := service.pending[playerID]
	delete(service.pending, playerID)
	if len(service.pending) == 0 {
		service.pending = nil
	}
	service.pendingMutex.Unlock()
	if found {
		service.releasePair(pending.transit)
	}
}

// pendingFor reports whether a player owns an active destination handoff.
func (service *Service) pendingFor(playerID int64, roomID int64) bool {
	service.pendingMutex.Lock()
	pending, found := service.pending[playerID]
	service.pendingMutex.Unlock()

	return found && pending.roomID == roomID
}

// entered applies the paired destination before room entity bootstrap.
func (service *Service) entered(ctx context.Context, payload roomentered.Payload) error {
	pending, found := service.consumePending(payload.PlayerID, payload.RoomID)
	if !found {
		return nil
	}
	active, found := service.runtime.Find(payload.RoomID)
	if !found {
		service.releasePair(pending.transit)
		return nil
	}
	target, found := active.FurnitureItem(pending.transit.Target.ID)
	if !found {
		service.releasePair(pending.transit)
		return service.fail(ctx, payload.RoomID, pending.transit, "target_removed")
	}
	_, err := active.TeleportUnit(payload.PlayerID, target.Point, target.Rotation, true)
	if err != nil {
		service.releasePair(pending.transit)
		return service.fail(ctx, payload.RoomID, pending.transit, "target_unavailable")
	}
	state := service.roomState(payload.RoomID)
	state.mutex.Lock()
	transit := pending.transit
	transit.Target = target
	transit.Phase = PhaseArrival
	transit.Deadline = service.now().Add(delayFor(transit))
	state.transits[payload.PlayerID] = transit
	state.mutex.Unlock()
	return nil
}

// publishStarted emits one accepted transition event.
func (service *Service) publishStarted(ctx context.Context, roomID int64, transit Transit) error {
	if service.events == nil {
		return nil
	}

	return service.events.Publish(ctx, bus.Event{Name: teleportstarted.Name, Payload: teleportstarted.Payload{
		PlayerID: transit.PlayerID, SourceItemID: transit.Source.ID, SourceRoomID: roomID,
		TargetItemID: transit.Target.ID, TargetRoomID: transit.TargetRoomID,
	}})
}

// fail clears control and emits a diagnostic transition event.
func (service *Service) fail(ctx context.Context, roomID int64, transit Transit, reason string) error {
	if active, found := service.runtime.Find(roomID); found {
		if unit, unitFound := active.Unit(transit.PlayerID); unitFound {
			_, _ = active.TeleportUnit(transit.PlayerID, unit.Position.Point, unit.BodyRotation, false)
		}
	}
	if service.events == nil {
		return nil
	}

	return service.events.Publish(ctx, bus.Event{Name: teleportfailed.Name, Payload: teleportfailed.Payload{
		PlayerID: transit.PlayerID, SourceItemID: transit.Source.ID, RoomID: roomID, Reason: reason,
	}})
}

// Register wires room cycle and lifecycle event integration.
func Register(lifecycle fx.Lifecycle, subscriber bus.Subscriber, runtime *roomlive.Registry, service *Service) error {
	runtime.AddCyclePublisher(service.Cycle)
	enteredSubscription, err := subscriber.Subscribe(roomentered.Name, bus.PriorityHigh, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(roomentered.Payload)
		if !ok {
			return nil
		}

		return service.entered(ctx, payload)
	})
	if err != nil {
		return err
	}
	disconnectedSubscription, err := subscriber.Subscribe(playerdisconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerdisconnected.Payload)
		if ok {
			service.clearPending(payload.PlayerID)
			service.cancelPlayer(ctx, payload.PlayerID)
		}

		return nil
	})
	if err != nil {
		enteredSubscription.Unsubscribe()
		return err
	}
	walkedSubscription, err := subscriber.Subscribe(furniturewalkedon.Name, bus.PriorityHigh, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(furniturewalkedon.Payload)
		if !ok {
			return nil
		}
		active, found := runtime.Find(payload.RoomID)
		if !found {
			return nil
		}
		item, found := active.FurnitureItem(payload.ItemID)
		if !found || item.Definition.InteractionType != "teleport_tile" {
			return nil
		}

		return service.Start(ctx, StartRequest{PlayerID: payload.PlayerID, Room: active, ItemID: payload.ItemID})
	})
	if err != nil {
		enteredSubscription.Unsubscribe()
		disconnectedSubscription.Unsubscribe()
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		enteredSubscription.Unsubscribe()
		disconnectedSubscription.Unsubscribe()
		walkedSubscription.Unsubscribe()
		return nil
	}})

	return nil
}
