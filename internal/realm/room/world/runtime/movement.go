package runtime

import (
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldindex "github.com/niflaot/pixels/internal/realm/room/world/runtime/furnitureindex"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// PlanMovement captures pathfinding input without mutating the world.
func (world *World) PlanMovement(playerID int64, goal grid.Point) (MovementPlan, error) {
	roomUnit, ok := world.units[playerID]
	if !ok {
		return MovementPlan{}, ErrUnitNotFound
	}

	return MovementPlan{
		Start: roomUnit.Position(), Goal: world.resolveSlotGoal(goal),
		Occupancy: world.occupancyExcept(playerID),
	}, nil
}

// FindPath resolves a path from a previously captured movement plan.
func (world *World) FindPath(plan MovementPlan) (worldpath.Path, error) {
	finder := worldpath.NewFinderWithOccupancy(world.resolver, world.rules, plan.Occupancy)

	return finder.Find(plan.Start, plan.Goal)
}

// ApplyMovement assigns a validated path to a unit.
func (world *World) ApplyMovement(playerID int64, roomPath worldpath.Path, exiting bool) error {
	control := worldunit.ControlNone
	if exiting {
		control = worldunit.ControlExitingRoom
	}

	return world.ApplyControlledMovement(playerID, roomPath, control)
}

// ApplyControlledMovement assigns a path with explicit server control.
func (world *World) ApplyControlledMovement(playerID int64, roomPath worldpath.Path, control worldunit.ControlKind) error {
	roomUnit, ok := world.units[playerID]
	if !ok {
		return ErrUnitNotFound
	}
	if roomUnit.Control() != worldunit.ControlNone && control == worldunit.ControlNone {
		return ErrUnitExiting
	}
	if err := roomPath.Validate(world.resolver); err != nil {
		return err
	}
	world.assignLinkedPath(playerID, roomPath, control)

	return nil
}

// assignLinkedPath assigns independent path cursors to one unit and its mounted counterpart.
func (world *World) assignLinkedPath(entityKey int64, roomPath worldpath.Path, control worldunit.ControlKind) {
	roomUnit := world.units[entityKey]
	world.releaseSlot(entityKey)
	roomUnit.SetPath(roomPath)
	if control != worldunit.ControlNone {
		roomUnit.SetControl(control)
	}
	linkedKey, linked := world.linkedKey(entityKey)
	if !linked {
		return
	}
	linkedUnit, found := world.units[linkedKey]
	if !found {
		world.unlinkMount(entityKey)
		return
	}
	world.releaseSlot(linkedKey)
	linkedUnit.SetPath(roomPath)
	if control != worldunit.ControlNone {
		linkedUnit.SetControl(control)
	}
}

// linkedKey returns the entity sharing authoritative movement with one mounted unit.
func (world *World) linkedKey(entityKey int64) (int64, bool) {
	return world.mounts.Linked(entityKey)
}

// TeleportUnit repositions one unit without pathfinding.
func (world *World) TeleportUnit(playerID int64, point grid.Point, rotation worldunit.Rotation, controlled bool, policies ...TeleportPolicy) (UnitSnapshot, error) {
	roomUnit, ok := world.units[playerID]
	if !ok {
		return UnitSnapshot{}, ErrUnitNotFound
	}
	if len(policies) > 0 && policies[0] == TeleportNear {
		return world.teleportNear(playerID, roomUnit, point, rotation, controlled)
	}
	section, err := world.resolver.TopSection(point)
	if err != nil {
		return UnitSnapshot{}, err
	}
	return world.repositionUnit(playerID, roomUnit, point, section.Z(), rotation, controlled), nil
}

// ApplyControlledStep assigns one authoritative adjacent movement step.
func (world *World) ApplyControlledStep(playerID int64, point grid.Point, control worldunit.ControlKind) error {
	_, ok := world.units[playerID]
	if !ok {
		return ErrUnitNotFound
	}
	section, err := world.resolver.TopSection(point)
	if err != nil || !world.rules.AllowsSection(section) {
		return worldpath.ErrInvalidGoal
	}
	step := worldpath.Step{Position: worldpath.Position{Point: point, Z: section.Z()}}
	return world.ApplyControlledMovement(playerID, worldpath.NewPath([]worldpath.Step{step}), control)
}

// FaceTo rotates a free-standing unit toward a target and clears movement.
func (world *World) FaceTo(playerID int64, target grid.Point) (UnitSnapshot, error) {
	roomUnit, ok := world.units[playerID]
	if !ok {
		return UnitSnapshot{}, ErrUnitNotFound
	}
	if roomUnit.Exiting() || roomUnit.Settled() {
		return unitSnapshot(playerID, roomUnit), nil
	}
	roomUnit.ClearPath()
	roomUnit.FaceToward(target)
	if linkedKey, linked := world.linkedKey(playerID); linked {
		if linkedUnit, found := world.units[linkedKey]; found {
			linkedUnit.ClearPath()
			linkedUnit.FaceToward(target)
		}
	}

	return unitSnapshot(playerID, roomUnit), nil
}

// Tick advances every moving unit once in stable player order.
func (world *World) Tick() []Movement {
	playerIDs := world.sortedPlayerIDs()
	movements := make([]Movement, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		roomUnit := world.units[playerID]
		if roomUnit.Moving() {
			if err := roomUnit.ValidatePath(world.resolver); err != nil {
				exited := roomUnit.Exiting()
				roomUnit.ClearPath()
				if section, sectionErr := world.nearestWalkableSection(roomUnit.Position()); sectionErr == nil {
					roomUnit.SetHeight(section.Z())
				}
				movements = append(movements, Movement{
					PlayerID: playerID, Unit: unitSnapshot(playerID, roomUnit),
					Settled: true, Exited: exited, ForcedExit: exited,
				})
				continue
			}
		}
		step, moved, settled := roomUnit.Advance()
		if !moved && !settled {
			continue
		}
		if settled {
			world.settleUnit(playerID, roomUnit)
		}
		exited := settled && roomUnit.Kind() == worldunit.KindPlayer && roomUnit.Position().Point == world.door.Point
		movements = append(movements, Movement{
			PlayerID: playerID, Unit: unitSnapshot(playerID, roomUnit), Step: step,
			Moved: moved, Settled: settled, Exited: exited,
			ForcedExit: exited && roomUnit.Exiting(),
		})
	}

	return movements
}

// resolveSlotGoal maps a furniture footprint target onto a usable slot.
func (world *World) resolveSlotGoal(goal grid.Point) grid.Point {
	return worldindex.ResolveSlotGoal(world.furniture, goal)
}

// settleUnit applies furniture sit or lay state at a unit's position.
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

// settleOnSection applies slot rotation and relative status height when available.
func (world *World) settleOnSection(playerID int64, roomUnit *worldunit.Unit, point grid.Point, slotStatus worldfurniture.SlotStatus, unitStatus string, section surface.Section) {
	if slot, found := world.slotAt(point, slotStatus); found {
		world.occupySlot(playerID, point)
		roomUnit.Settle(unitStatus, heightValue(slot.Z-section.Z()), slot.BodyRotation, slot.BodyRotation)
		return
	}
	roomUnit.Settle(unitStatus, heightValue(section.Z()), roomUnit.BodyRotation(), roomUnit.HeadRotation())
}

// slotAt finds a furniture slot at a tile with a matching status.
func (world *World) slotAt(point grid.Point, status worldfurniture.SlotStatus) (worldfurniture.Slot, bool) {
	for _, item := range world.furniture {
		for _, slot := range worldfurniture.Slots(item) {
			if slot.Point == point && slot.Status == status {
				return slot, true
			}
		}
	}

	return worldfurniture.Slot{}, false
}

// occupancyExcept returns occupied and reserved positions except one player.
func (world *World) occupancyExcept(playerID int64) worldpath.Occupancy {
	positions := make([]worldpath.Position, 0, len(world.units)*2)
	linkedKey, linked := world.linkedKey(playerID)
	for occupantID, roomUnit := range world.units {
		if occupantID == playerID || linked && occupantID == linkedKey {
			continue
		}
		positions = append(positions, roomUnit.Position())
		if goal, ok := roomUnit.Goal(); ok {
			positions = append(positions, goal)
		}
	}

	return worldpath.NewOccupancy(positions)
}

// heightValue formats a grid height for a unit status value.
func heightValue(height grid.Height) string {
	return height.String()
}
