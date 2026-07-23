// Package furniture executes WIRED furniture mutations through authoritative room placement.
package furniture

import (
	"context"
	"strconv"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	roomitems "github.com/niflaot/pixels/internal/realm/room/world/items"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	netconn "github.com/niflaot/pixels/networking/connection"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
)

// Manager combines authoritative furniture records and state updates.
type Manager interface {
	furnitureservice.Manager
	furnitureservice.StateUpdater
}

// Service executes furniture effects.
type Service struct {
	// rooms resolves active rooms.
	rooms *roomlive.Registry
	// furniture persists furniture placement and state.
	furniture Manager
	// connections broadcasts committed projections.
	connections *netconn.Registry
}

// New creates a WIRED furniture effect service.
func New(rooms *roomlive.Registry, furniture Manager, connections *netconn.Registry) *Service {
	return &Service{rooms: rooms, furniture: furniture, connections: connections}
}

// ExecuteFurniture executes one validated furniture mutation.
func (service *Service) ExecuteFurniture(ctx context.Context, operation effect.FurnitureOperation, node *configuration.Node, event trigger.Event) (effect.Result, error) {
	active, found := service.rooms.Find(event.RoomID)
	if !found {
		return effect.Result{Status: effect.Skipped}, nil
	}
	var result effect.Result
	result.Status = effect.Skipped
	for _, target := range node.Targets {
		mutation, err := service.executeTarget(ctx, active, operation, node, event, target)
		if err != nil {
			return effect.Result{Status: effect.Blocked}, err
		}
		result.Derived = appendCollision(result.Derived, mutation)
		if mutation.changed {
			result.Status = effect.Applied
			derived := trigger.Event{Kind: trigger.StateChanged, RoomID: event.RoomID, ActorKind: event.ActorKind, ActorID: event.ActorID, PlayerID: event.PlayerID, SourceItem: target.ItemID}
			if item, exists := active.FurnitureItem(target.ItemID); exists {
				derived.SourceSprite = int32(item.Definition.SpriteID)
			}
			result.Derived = append(result.Derived, derived)
		} else if mutation.collided {
			result.Status = effect.Applied
		}
	}
	if node.SelectionMode == 3 && event.SourceItem > 0 && !containsTarget(node.Targets, event.SourceItem) {
		mutation, err := service.executeTarget(ctx, active, operation, node, event, record.Target{ItemID: event.SourceItem})
		if err != nil {
			return effect.Result{Status: effect.Blocked}, err
		}
		result.Derived = appendCollision(result.Derived, mutation)
		if mutation.changed {
			result.Status = effect.Applied
			result.Derived = append(result.Derived, trigger.Event{Kind: trigger.StateChanged, RoomID: event.RoomID, ActorKind: event.ActorKind, ActorID: event.ActorID, PlayerID: event.PlayerID, SourceItem: event.SourceItem, SourceSprite: event.SourceSprite})
		} else if mutation.collided {
			result.Status = effect.Applied
		}
	}
	return result, nil
}

// mutation stores one target's committed change or collision outcome.
type mutation struct {
	// changed reports a committed furniture mutation.
	changed bool
	// collided reports that a unit prevented the requested movement.
	collided bool
	// collision stores the derived unit collision context.
	collision trigger.Event
}

// executeTarget applies one furniture operation to a resolved target.
func (service *Service) executeTarget(ctx context.Context, active *roomlive.Room, operation effect.FurnitureOperation, node *configuration.Node, event trigger.Event, target record.Target) (mutation, error) {
	switch operation {
	case effect.ToggleState, effect.ToggleRandomState:
		changed, err := service.toggle(ctx, active, target.ItemID, operation == effect.ToggleRandomState, event.ID)
		return mutation{changed: changed}, err
	case effect.MatchSnapshot:
		changed, err := service.restore(ctx, active, target, event.PlayerID)
		return mutation{changed: changed}, err
	default:
		return service.move(ctx, active, target.ItemID, operation, node, event)
	}
}

// appendCollision appends a collision event only when a movement encountered a unit.
func appendCollision(events []trigger.Event, result mutation) []trigger.Event {
	if !result.collided {
		return events
	}

	return append(events, result.collision)
}

// containsTarget reports whether an explicit selection already contains an item.
func containsTarget(targets []record.Target, itemID int64) bool {
	for _, target := range targets {
		if target.ItemID == itemID {
			return true
		}
	}
	return false
}

// toggle advances or deterministically randomizes one furniture state.
func (service *Service) toggle(ctx context.Context, active *roomlive.Room, itemID int64, random bool, eventID uint64) (bool, error) {
	item, found := active.FurnitureItem(itemID)
	if !found || item.Definition.InteractionModesCount <= 1 {
		return false, nil
	}
	current, err := strconv.Atoi(item.ExtraData)
	if err != nil || current < 0 || current >= item.Definition.InteractionModesCount {
		current = 0
	}
	next := (current + 1) % item.Definition.InteractionModesCount
	if random && item.Definition.InteractionModesCount > 2 {
		next = int((eventID+uint64(itemID))%uint64(item.Definition.InteractionModesCount-1)) + 1
		if next == current {
			next = (next + 1) % item.Definition.InteractionModesCount
		}
	}
	value := strconv.Itoa(next)
	if _, err = service.furniture.UpdateState(ctx, furnitureservice.StateParams{ItemID: itemID, RoomID: active.ID(), Expected: item.ExtraData, Next: value}); err != nil {
		return false, err
	}
	updated, err := active.UpdateFurnitureState(itemID, value, true)
	if err != nil {
		return false, err
	}
	return true, service.broadcast(ctx, active, updated)
}

// restore applies a captured snapshot through normal placement and state validation.
func (service *Service) restore(ctx context.Context, active *roomlive.Room, target record.Target, actorID int64) (bool, error) {
	if !target.Snapshot.Present {
		return false, nil
	}
	item, found := active.FurnitureItem(target.ItemID)
	if !found {
		return false, nil
	}
	point, valid := grid.NewPoint(target.Snapshot.X, target.Snapshot.Y)
	if !valid {
		return false, nil
	}
	rotation := worldunit.Rotation(target.Snapshot.Rotation)
	changed, err := service.place(ctx, active, item, point, rotation, actorID)
	if err != nil || !changed {
		return changed, err
	}
	current, found := active.FurnitureItem(item.ID)
	if found && current.ExtraData != target.Snapshot.State {
		if _, err = service.furniture.UpdateState(ctx, furnitureservice.StateParams{ItemID: item.ID, RoomID: active.ID(), Expected: current.ExtraData, Next: target.Snapshot.State}); err != nil {
			return false, err
		}
		current, err = active.UpdateFurnitureState(item.ID, target.Snapshot.State, true)
		if err != nil {
			return false, err
		}
		return true, service.broadcast(ctx, active, current)
	}
	return true, nil
}

// broadcast projects one committed furniture mutation.
func (service *Service) broadcast(ctx context.Context, active *roomlive.Room, item worldfurniture.Item) error {
	packet, err := outupdate.Encode(outupdate.FloorItem{ID: item.ID, SpriteID: item.Definition.SpriteID, X: int(item.Point.X), Y: int(item.Point.Y), Rotation: int(item.Rotation), Z: item.Z.String(), ExtraHeight: item.Top().String(), ExtraData: item.ExtraData, OwnerID: item.OwnerPlayerID})
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}

// durableRotation maps world rotation to persistence rotation.
func durableRotation(rotation worldunit.Rotation) furnituremodel.Rotation {
	return furnituremodel.Rotation(rotation)
}

// resolveWorldItem validates a destination against current surfaces and footprints.
func (service *Service) resolveWorldItem(ctx context.Context, active *roomlive.Room, itemID int64, point grid.Point, rotation worldunit.Rotation) (worldfurniture.Item, bool, error) {
	durable, found, err := service.furniture.FindItemByID(ctx, itemID)
	if err != nil || !found {
		return worldfurniture.Item{}, false, err
	}
	resolved, _, err := roomitems.ResolveWorldItem(ctx, active, service.furniture, durable, int(point.X), int(point.Y), durableRotation(rotation))
	return resolved, err == nil, err
}
