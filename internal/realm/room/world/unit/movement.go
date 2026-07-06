package unit

import (
	"strconv"

	"github.com/niflaot/pixels/internal/realm/room/world/path"
)

// SetPath stores a new pending movement path.
func (unit *Unit) SetPath(roomPath path.Path) {
	steps := roomPath.Steps()
	unit.steps = steps
	unit.hasGoal = len(steps) > 0
	if unit.hasGoal {
		unit.goal = steps[len(steps)-1].Position
		unit.setMoveStatus(steps[0].Position)
	}
}

// ClearPath clears pending movement.
func (unit *Unit) ClearPath() {
	unit.steps = nil
	unit.hasGoal = false
	unit.statuses.clear(StatusMove)
}

// Moving reports whether the unit has pending steps.
func (unit *Unit) Moving() bool {
	return len(unit.steps) > 0
}

// Goal returns the active movement goal.
func (unit *Unit) Goal() (path.Position, bool) {
	return unit.goal, unit.hasGoal
}

// PendingSteps returns the number of pending movement steps.
func (unit *Unit) PendingSteps() int {
	return len(unit.steps)
}

// Advance moves the unit by one pending step.
func (unit *Unit) Advance() (path.Step, bool) {
	if len(unit.steps) == 0 {
		return path.Step{}, false
	}

	step := unit.steps[0]
	unit.steps = unit.steps[1:]
	unit.previous = unit.position
	unit.position = step.Position
	unit.rotateToward(step.Position)
	if len(unit.steps) == 0 {
		unit.hasGoal = false
		unit.statuses.clear(StatusMove)
	} else {
		unit.setMoveStatus(unit.steps[0].Position)
	}

	return step, true
}

// setMoveStatus stores the next movement status.
func (unit *Unit) setMoveStatus(position path.Position) {
	value := strconv.Itoa(int(position.Point.X)) + "," +
		strconv.Itoa(int(position.Point.Y)) + "," +
		strconv.Itoa(int(position.Z))
	unit.statuses.set(StatusMove, value)
}

// rotateToward rotates the unit toward a position.
func (unit *Unit) rotateToward(position path.Position) {
	rotation := RotationBetween(unit.previous.Point.X, unit.previous.Point.Y, position.Point.X, position.Point.Y)
	unit.body = rotation
	unit.head = rotation
}
