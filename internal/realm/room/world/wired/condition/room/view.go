// Package room adapts active room state to WIRED condition facts.
package room

import (
	socialgroup "github.com/niflaot/pixels/internal/realm/group"
	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/condition"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// Provider resolves active-room condition views.
type Provider struct {
	// rooms resolves active rooms.
	rooms *roomlive.Registry
	// games resolves ephemeral team state.
	games *game.Service
	// achievements resolves equipped badge snapshots.
	achievements *playerachievement.Service
	// groups resolves room-linked social membership.
	groups *socialgroup.Service
}

// New creates an active-room view provider.
func New(rooms *roomlive.Registry, games *game.Service, achievements *playerachievement.Service, groups *socialgroup.Service) *Provider {
	return &Provider{rooms: rooms, games: games, achievements: achievements, groups: groups}
}

// View returns one active-room condition adapter.
func (provider *Provider) View(roomID int64) (condition.View, bool) {
	active, found := provider.rooms.Find(roomID)
	if !found {
		return nil, false
	}
	return &View{active: active, games: provider.games, achievements: provider.achievements, groups: provider.groups}, true
}

// View reads authoritative live room facts.
type View struct {
	// active stores the active room.
	active *roomlive.Room
	// games resolves room game state.
	games *game.Service
	// achievements resolves equipped badge snapshots.
	achievements *playerachievement.Service
	// groups resolves room-linked social membership.
	groups *socialgroup.Service
}

// UserCount returns player occupancy excluding bot and pet units.
func (view *View) UserCount() int { return view.active.Occupancy().Count }

// UnitOn reports whether any room unit occupies a furniture footprint.
func (view *View) UnitOn(itemID int64) (bool, error) {
	item, found := view.active.FurnitureItem(itemID)
	if !found {
		return false, nil
	}
	_, occupied := view.active.UnitInFurnitureFootprint(item)

	return occupied, nil
}

// ActorOn reports whether the event actor occupies a furniture footprint.
func (view *View) ActorOn(event trigger.Event, itemID int64) (bool, bool, error) {
	unit, found := view.active.UnitMotion(event.ActorID)
	if !found {
		return false, false, nil
	}
	item, found := view.active.FurnitureItem(itemID)
	if !found {
		return false, true, nil
	}
	return footprintContains(item, unit.Position.Point), true, nil
}

// Stacked reports whether another furniture item sits on a base item.
func (view *View) Stacked(itemID int64) (bool, error) {
	base, found := view.active.FurnitureItem(itemID)
	if !found {
		return false, nil
	}
	for _, candidate := range view.active.FurnitureItems() {
		if candidate.ID != base.ID && candidate.Z >= base.Top() && footprintsOverlap(base, candidate) {
			return true, nil
		}
	}
	return false, nil
}

// SnapshotMatches compares live furniture with captured fields and flags.
func (view *View) SnapshotMatches(itemID int64, snapshot record.Snapshot, flags []int32) (bool, error) {
	item, found := view.active.FurnitureItem(itemID)
	if !found || !snapshot.Present {
		return false, nil
	}
	state, position, rotation := enabled(flags, 0), enabled(flags, 1), enabled(flags, 2)
	if state && item.ExtraData != snapshot.State {
		return false, nil
	}
	if position && (int(item.Point.X) != snapshot.X || int(item.Point.Y) != snapshot.Y || item.Z.Units() != snapshot.Z) {
		return false, nil
	}
	return !rotation || int(item.Rotation) == snapshot.Rotation, nil
}

// ActorTeam reports player team membership.
func (view *View) ActorTeam(playerID int64, team int32) (bool, bool, error) {
	if playerID <= 0 || view.games == nil {
		return false, false, nil
	}
	actual, found := view.games.Team(view.active.ID(), playerID)
	return found && actual == team, true, nil
}

// ActorGroup reports whether the event actor belongs to the room's social group.
func (view *View) ActorGroup(playerID int64) (bool, bool, error) {
	if view.groups == nil || playerID <= 0 {
		return false, false, nil
	}
	pass, loaded := view.groups.IsRoomMember(view.active.ID(), playerID)
	return pass, loaded, nil
}

// WearingBadge reports whether the event actor has the configured badge equipped.
func (view *View) WearingBadge(playerID int64, code string) (bool, bool, error) {
	if view.achievements == nil || playerID <= 0 || code == "" {
		return false, false, nil
	}
	pass, loaded := view.achievements.Wearing(playerID, code)
	return pass, loaded, nil
}

// WearingEffect reports a player's current room effect.
func (view *View) WearingEffect(playerID int64, effectID int32) (bool, bool, error) {
	unit, found := view.active.Unit(playerID)
	return found && unit.ActiveEffectID == effectID, found, nil
}

// HasHanditem reports a player's current room hand item.
func (view *View) HasHanditem(playerID int64, itemID int32) (bool, bool, error) {
	unit, found := view.active.Unit(playerID)
	return found && unit.HandItem == itemID, found, nil
}

// footprintContains reports whether a point occupies a furniture footprint.
func footprintContains(item furniture.Item, point grid.Point) bool {
	for _, candidate := range furniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
		if candidate == point {
			return true
		}
	}
	return false
}

// footprintsOverlap reports whether two furniture footprints share a tile.
func footprintsOverlap(left furniture.Item, right furniture.Item) bool {
	for _, leftPoint := range furniture.Footprint(left.Point, left.Definition.Width, left.Definition.Length, left.Rotation) {
		for _, rightPoint := range furniture.Footprint(right.Point, right.Definition.Width, right.Definition.Length, right.Rotation) {
			if leftPoint == rightPoint {
				return true
			}
		}
	}
	return false
}

// enabled reports whether an editor flag is enabled and defaults missing flags on.
func enabled(flags []int32, index int) bool {
	return index >= len(flags) || flags[index] != 0
}
