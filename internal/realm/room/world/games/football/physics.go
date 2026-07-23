// Package football implements authoritative football physics and scoring.
package football

// Vector stores one of Habbo's eight movement directions.
type Vector struct {
	// X stores horizontal movement.
	X int
	// Y stores vertical movement.
	Y int
}

// Direction returns the canonical vector for a Habbo rotation.
func Direction(rotation uint8) Vector {
	directions := [...]Vector{{X: 0, Y: -1}, {X: 1, Y: -1}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}, {X: -1, Y: 1}, {X: -1, Y: 0}, {X: -1, Y: -1}}
	return directions[rotation%8]
}

// Reflect returns the eight-direction rebound after axis collisions.
func Reflect(rotation uint8, blockedX bool, blockedY bool) uint8 {
	vector := Direction(rotation)
	if blockedX {
		vector.X = -vector.X
	}
	if blockedY {
		vector.Y = -vector.Y
	}
	for candidate := uint8(0); candidate < 8; candidate++ {
		if Direction(candidate) == vector {
			return candidate
		}
	}
	return rotation % 8
}

// Rebounds returns the preferred directions after a blocked movement.
func Rebounds(rotation uint8) [3]uint8 {
	switch rotation % 8 {
	case 1:
		return [3]uint8{7, 3, 5}
	case 3:
		return [3]uint8{5, 1, 7}
	case 5:
		return [3]uint8{3, 7, 1}
	case 7:
		return [3]uint8{1, 5, 3}
	default:
		reverse := (rotation + 4) % 8
		return [3]uint8{reverse, reverse, reverse}
	}
}

// GoalScores reports whether a ball entered through the goal's scoring face.
func GoalScores(movementRotation uint8, goalRotation uint8) bool {
	scoringDirection := (goalRotation + 4) % 8
	difference := (movementRotation + 8 - scoringDirection) % 8
	return difference == 0 || difference == 1 || difference == 7
}

// Scoreboard stores one football counter value.
type Scoreboard struct {
	// Value stores zero through 99.
	Value int
}

// Add changes the score and wraps through zero to 99.
func (scoreboard *Scoreboard) Add(delta int) int {
	value := (scoreboard.Value + delta) % 100
	if value < 0 {
		value += 100
	}
	scoreboard.Value = value
	return value
}

// Reset clears the scoreboard.
func (scoreboard *Scoreboard) Reset() { scoreboard.Value = 0 }
