package unit

import (
	"strconv"

	"github.com/niflaot/pixels/internal/realm/room/world/path"
)

// SetPath stores a new pending movement path.
func (unit *Unit) SetPath(roomPath path.Path) {
	steps := roomPath.Steps()
	unit.activePath = roomPath
	unit.steps = steps
	unit.hasGoal = len(steps) > 0
	unit.settling = false
	unit.statuses.clear(StatusSit)
	unit.statuses.clear(StatusLay)
	if unit.hasGoal {
		unit.goal = steps[len(steps)-1].Position
		unit.setMoveStatus(steps[0].Position)
	}
}

// ClearPath clears pending movement.
func (unit *Unit) ClearPath() {
	unit.activePath = path.Path{}
	unit.steps = nil
	unit.hasGoal = false
	unit.settling = false
	unit.statuses.clear(StatusMove)
}

// ValidatePath reports whether the unit's active path still matches the current world state.
func (unit *Unit) ValidatePath(world path.World) error {
	if len(unit.steps) == 0 && !unit.settling {
		return nil
	}

	return unit.activePath.Validate(world)
}

// Settle finalizes a unit's landed status and rotation, replacing any pending movement status.
func (unit *Unit) Settle(status string, value string, body Rotation, head Rotation) {
	unit.statuses.clear(StatusMove)
	unit.statuses.set(status, value)
	unit.body = body
	unit.head = head
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

// MarkExiting prevents client movement from replacing a server-controlled exit path.
func (unit *Unit) MarkExiting() {
	unit.control = ControlExitingRoom
}

// Exiting reports whether the unit is following a server-controlled exit path.
func (unit *Unit) Exiting() bool {
	return unit.control == ControlExitingRoom
}

// SetControl assigns server control over unit movement.
func (unit *Unit) SetControl(control ControlKind) {
	unit.control = control
}

// Control returns the active server movement control.
func (unit *Unit) Control() ControlKind {
	return unit.control
}

// Reposition moves a unit immediately and clears transient posture and movement state.
func (unit *Unit) Reposition(position path.Position, rotation Rotation) {
	unit.previous = unit.position
	unit.position = position
	unit.body = rotation
	unit.head = rotation
	unit.ClearPath()
	unit.statuses.clear(StatusSit)
	unit.statuses.clear(StatusLay)
}

// Advance moves the unit by one pending step.
func (unit *Unit) Advance() (path.Step, bool, bool) {
	if len(unit.steps) == 0 {
		if !unit.settling {
			return path.Step{}, false, false
		}
		unit.settling = false
		unit.statuses.clear(StatusMove)

		return path.Step{}, false, true
	}

	step := unit.steps[0]
	unit.steps = unit.steps[1:]
	unit.previous = unit.position
	unit.position = step.Position
	unit.rotateToward(step.Position)
	unit.setMoveStatus(step.Position)
	if len(unit.steps) == 0 {
		unit.hasGoal = false
		unit.settling = true
	}

	return step, true, false
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
