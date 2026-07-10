package live

import (
	"strconv"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// MoveTo sets a unit movement goal. Goals inside a slotted furniture item's footprint snap to the
// item's matching slot tile, so clicking anywhere on a bed walks the unit to its anchor slot.
func (room *Room) MoveTo(playerID int64, goal grid.Point) (worldpath.Path, error) {
	return room.moveTo(playerID, goal, false)
}

// ExitToDoor starts a server-controlled path to the room door.
func (room *Room) ExitToDoor(playerID int64) (bool, error) {
	room.mutex.RLock()
	if room.world == nil {
		room.mutex.RUnlock()
		return false, ErrWorldNotLoaded
	}
	door := room.world.door.Point
	room.mutex.RUnlock()

	roomPath, err := room.moveTo(playerID, door, true)
	if err != nil {
		return false, err
	}

	return roomPath.Len() > 0, nil
}

// moveTo sets client or server-controlled unit movement.
func (room *Room) moveTo(playerID int64, goal grid.Point, exiting bool) (worldpath.Path, error) {
	runtime, start, occupancy, goal, err := room.movementSnapshot(playerID, goal)
	if err != nil {
		return worldpath.Path{}, err
	}

	finder := worldpath.NewFinderWithOccupancy(runtime.resolver, runtime.rules, occupancy)
	roomPath, err := finder.Find(start, goal)
	if err != nil {
		return worldpath.Path{}, err
	}
	if len(roomPath.Steps()) == 0 {
		// The goal resolved to the unit's current tile (e.g. re-clicking the bed you're already
		// lying on). Treat it as a no-op: releasing the slot and clearing sit/lay here would silently
		// desync World.slotOccupants from the unit's still-settled status, since a zero-step path never
		// advances on tick and so never gets broadcast (see Room.Tick/Unit.Advance).
		return roomPath, nil
	}

	room.mutex.Lock()
	defer room.mutex.Unlock()

	if room.world != runtime {
		return worldpath.Path{}, worldpath.ErrInvalidPath
	}
	roomUnit, ok := room.world.units[playerID]
	if !ok {
		return worldpath.Path{}, ErrUnitNotFound
	}
	if roomUnit.Exiting() && !exiting {
		return worldpath.Path{}, ErrUnitExiting
	}
	if err := roomPath.Validate(room.world.resolver); err != nil {
		return worldpath.Path{}, err
	}
	room.world.releaseSlot(playerID)
	roomUnit.SetPath(roomPath)
	if exiting {
		roomUnit.MarkExiting()
	}

	return roomPath, nil
}

// FaceTo rotates a unit toward a target point and clears pending movement. Units settled on a sit
// or lay slot keep their forced slot rotation and stay put instead of spinning in place.
func (room *Room) FaceTo(playerID int64, target grid.Point) (UnitSnapshot, error) {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	if room.world == nil {
		return UnitSnapshot{}, ErrWorldNotLoaded
	}
	roomUnit, ok := room.world.units[playerID]
	if !ok {
		return UnitSnapshot{}, ErrUnitNotFound
	}
	if roomUnit.Exiting() {
		return unitSnapshot(playerID, roomUnit), nil
	}
	if roomUnit.Settled() {
		return unitSnapshot(playerID, roomUnit), nil
	}
	roomUnit.ClearPath()
	roomUnit.FaceToward(target)

	return unitSnapshot(playerID, roomUnit), nil
}

// Tick advances room world movement once.
func (room *Room) Tick() []Movement {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	if room.world == nil {
		return nil
	}

	playerIDs := room.world.sortedPlayerIDs()
	movements := make([]Movement, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		roomUnit := room.world.units[playerID]
		if roomUnit.Moving() {
			if err := roomUnit.ValidatePath(room.world.resolver); err != nil {
				exited := roomUnit.Exiting()
				roomUnit.ClearPath()
				if section, sectionErr := room.world.resolver.TopSection(roomUnit.Position().Point); sectionErr == nil {
					roomUnit.SetHeight(section.Z())
				}
				movements = append(movements, Movement{
					PlayerID: playerID, Unit: unitSnapshot(playerID, roomUnit), Settled: true,
					Exited: exited, ForcedExit: exited,
				})

				continue
			}
		}
		step, moved, settled := roomUnit.Advance()
		if !moved && !settled {
			continue
		}
		if settled {
			room.world.settleUnit(playerID, roomUnit)
		}
		exited := settled && roomUnit.Position().Point == room.world.door.Point
		forcedExit := exited && roomUnit.Exiting()
		movements = append(movements, Movement{
			PlayerID:   playerID,
			Unit:       unitSnapshot(playerID, roomUnit),
			Step:       step,
			Moved:      moved,
			Settled:    settled,
			Exited:     exited,
			ForcedExit: forcedExit,
		})
	}

	return movements
}

// movementSnapshot returns data needed to calculate movement outside the room lock, snapping the
// goal onto a slotted furniture item's anchor slot when it targets the item's footprint.
func (room *Room) movementSnapshot(playerID int64, goal grid.Point) (*World, worldpath.Position, worldpath.Occupancy, grid.Point, error) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	if room.world == nil {
		return nil, worldpath.Position{}, worldpath.Occupancy{}, goal, ErrWorldNotLoaded
	}
	roomUnit, ok := room.world.units[playerID]
	if !ok {
		return nil, worldpath.Position{}, worldpath.Occupancy{}, goal, ErrUnitNotFound
	}

	return room.world, roomUnit.Position(), room.world.occupancyExcept(playerID), room.world.resolveSlotGoal(goal), nil
}

// settleUnit applies a sit or lay status when a unit lands on a seat or lay section.
func (world *World) settleUnit(playerID int64, roomUnit *worldunit.Unit) {
	position := roomUnit.Position()
	column, err := world.resolver.Column(position.Point)
	if err != nil {
		return
	}
	section, ok := column.SectionAt(position.Z)
	if !ok {
		return
	}

	switch section.State() {
	case surface.StateSit:
		world.settleOnSection(playerID, roomUnit, position.Point, worldfurniture.SlotStatusSit, worldunit.StatusSit, section)
	case surface.StateLay:
		world.settleOnSection(playerID, roomUnit, position.Point, worldfurniture.SlotStatusLay, worldunit.StatusLay, section)
	}
}

// settleOnSection settles a unit using a placed furniture item's slot rotation when available, falling
// back to the unit's arrival rotation for sit/lay sections with no matching slot metadata. The slot
// status carries the height offset between the slot top and the section the unit stands on, matching
// the real protocol where a floor chair reports unit z 0 with sit value equal to the seat height.
func (world *World) settleOnSection(playerID int64, roomUnit *worldunit.Unit, point grid.Point, slotStatus worldfurniture.SlotStatus, unitStatus string, section surface.Section) {
	if slot, found := world.slotAt(point, slotStatus); found {
		world.occupySlot(playerID, point)
		roomUnit.Settle(unitStatus, heightValue(slot.Z-section.Z()), slot.BodyRotation, slot.BodyRotation)

		return
	}

	roomUnit.Settle(unitStatus, heightValue(section.Z()), roomUnit.BodyRotation(), roomUnit.HeadRotation())
}

// heightValue formats a grid height for a unit status value.
func heightValue(height grid.Height) string {
	return strconv.Itoa(int(height))
}
