package live

import (
	"time"

	roomtask "github.com/niflaot/pixels/internal/realm/room/runtime/live/task"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// UnitMotion returns allocation-free movement state for one room-owned cycle.
func (room *Room) UnitMotion(entityKey int64) (UnitSnapshot, bool) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return UnitSnapshot{}, false
	}
	return room.world.UnitMotion(entityKey)
}

// MoveTo sets a unit movement goal.
func (room *Room) MoveTo(playerID int64, goal grid.Point) (worldpath.Path, error) {
	return room.moveTo(playerID, goal, worldunit.ControlNone)
}

// MoveControlled sets a unit movement goal reserved for server behavior.
func (room *Room) MoveControlled(playerID int64, goal grid.Point, control worldunit.ControlKind) (worldpath.Path, error) {
	return room.moveTo(playerID, goal, control)
}

// StepControlled assigns one authoritative adjacent movement step.
func (room *Room) StepControlled(playerID int64, goal grid.Point, control worldunit.ControlKind) error {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return ErrWorldNotLoaded
	}

	return room.world.ApplyControlledStep(playerID, goal, control)
}

// StepControlledOntoInteraction assigns one authoritative step onto an adjacent interaction tile.
func (room *Room) StepControlledOntoInteraction(playerID int64, goal grid.Point, control worldunit.ControlKind) error {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return ErrWorldNotLoaded
	}

	return room.world.ApplyControlledInteractionStep(playerID, goal, control)
}

// StepControlledFromInteraction assigns one authoritative step away from an adjacent interaction.
func (room *Room) StepControlledFromInteraction(playerID int64, goal grid.Point, control worldunit.ControlKind) error {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return ErrWorldNotLoaded
	}

	return room.world.ApplyControlledInteractionStep(playerID, goal, control)
}

// ExitToDoor starts a server-controlled path to the room door.
func (room *Room) ExitToDoor(playerID int64) (bool, error) {
	room.mutex.RLock()
	if room.world == nil {
		room.mutex.RUnlock()
		return false, ErrWorldNotLoaded
	}
	door := room.world.Door().Point
	room.mutex.RUnlock()
	roomPath, err := room.moveTo(playerID, door, worldunit.ControlExitingRoom)
	if err != nil {
		return false, err
	}

	return roomPath.Len() > 0, nil
}

// moveTo plans outside the room lock and applies against the same loaded world.
func (room *Room) moveTo(playerID int64, goal grid.Point, control worldunit.ControlKind) (worldpath.Path, error) {
	room.mutex.RLock()
	runtime := room.world
	if runtime == nil {
		room.mutex.RUnlock()
		return worldpath.Path{}, ErrWorldNotLoaded
	}
	plan, err := runtime.PlanMovement(playerID, goal)
	room.mutex.RUnlock()
	if err != nil {
		return worldpath.Path{}, err
	}
	roomPath, err := runtime.FindPath(plan)
	if err != nil {
		return worldpath.Path{}, err
	}
	if roomPath.Len() == 0 {
		return roomPath, nil
	}
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world != runtime {
		return worldpath.Path{}, worldpath.ErrInvalidPath
	}
	if err := runtime.ApplyControlledMovement(playerID, roomPath, control); err != nil {
		return worldpath.Path{}, err
	}

	return roomPath, nil
}

// StopMovement discards future steps while allowing Nitro's current step to settle.
func (room *Room) StopMovement(playerID int64) (bool, error) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return false, ErrWorldNotLoaded
	}

	return room.world.StopMovement(playerID)
}

// FaceTo rotates a unit toward a target point and clears pending movement.
func (room *Room) FaceTo(playerID int64, target grid.Point) (UnitSnapshot, error) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return UnitSnapshot{}, ErrWorldNotLoaded
	}

	return room.world.FaceTo(playerID, target)
}

// SetHandItem replaces one unit's carried hand item.
func (room *Room) SetHandItem(playerID int64, itemID int32) (UnitSnapshot, error) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return UnitSnapshot{}, ErrWorldNotLoaded
	}

	return room.world.SetHandItem(playerID, itemID)
}

// ReleaseUnitControl clears one unit's server-owned movement workflow.
func (room *Room) ReleaseUnitControl(playerID int64) (UnitSnapshot, error) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return UnitSnapshot{}, ErrWorldNotLoaded
	}

	return room.world.ReleaseControl(playerID)
}

// SetUnitControl assigns one server-owned movement workflow.
func (room *Room) SetUnitControl(playerID int64, control worldunit.ControlKind) (UnitSnapshot, error) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return UnitSnapshot{}, ErrWorldNotLoaded
	}

	return room.world.SetUnitControl(playerID, control)
}

// CanChangeFurnitureHeight reports whether an item has no furniture stacked above it.
func (room *Room) CanChangeFurnitureHeight(itemID int64) bool {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return false
	}

	return room.world.CanChangeFurnitureHeight(itemID)
}

// ResettleFurnitureUnits updates units standing over one changed furniture item.
func (room *Room) ResettleFurnitureUnits(itemID int64) []UnitSnapshot {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return nil
	}

	return room.world.ResettleFurnitureUnits(itemID)
}

// Tick advances room world movement once.
func (room *Room) Tick() []Movement {
	room.mutex.Lock()
	if room.world == nil {
		room.mutex.Unlock()
		room.runDueTasks(time.Now())
		return nil
	}
	movements := room.world.Tick()
	room.mutex.Unlock()
	room.runDueTasks(time.Now())

	return movements
}

// Schedule queues independent room-owned work after a delay.
func (room *Room) Schedule(after time.Duration, run func(time.Time)) {
	room.tasks.Schedule(time.Now().Add(after), run)
}

// ScheduleReplacing queues work after replacing the same non-zero key.
func (room *Room) ScheduleReplacing(key roomtask.Key, after time.Duration, run func(time.Time)) {
	room.tasks.Replace(key, time.Now().Add(after), run)
}

// RunScheduled executes work due at the supplied time.
func (room *Room) RunScheduled(now time.Time) {
	room.runDueTasks(now)
}

// runDueTasks executes callbacks after queue locks have been released.
func (room *Room) runDueTasks(now time.Time) {
	for _, run := range room.tasks.Due(now) {
		run(now)
	}
}

// TryLockInteraction acquires one furniture cooldown until a deadline.
func (room *Room) TryLockInteraction(itemID int64, until time.Time) bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if deadline, found := room.interactionLocks[itemID]; found && deadline.After(time.Now()) {
		return false
	}
	if room.interactionLocks == nil {
		room.interactionLocks = make(map[int64]time.Time)
	}
	room.interactionLocks[itemID] = until

	return true
}

// UnlockInteraction releases one furniture cooldown.
func (room *Room) UnlockInteraction(itemID int64) {
	room.mutex.Lock()
	delete(room.interactionLocks, itemID)
	room.mutex.Unlock()
}
