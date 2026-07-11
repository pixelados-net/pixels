package live

import (
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

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

// FaceTo rotates a unit toward a target point and clears pending movement.
func (room *Room) FaceTo(playerID int64, target grid.Point) (UnitSnapshot, error) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return UnitSnapshot{}, ErrWorldNotLoaded
	}

	return room.world.FaceTo(playerID, target)
}

// Tick advances room world movement once.
func (room *Room) Tick() []Movement {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return nil
	}

	return room.world.Tick()
}
