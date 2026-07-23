package football

import "testing"

// TestReflect covers every direction against vertical and horizontal walls.
func TestReflect(t *testing.T) {
	tests := []struct {
		direction uint8
		blockX    bool
		blockY    bool
		want      uint8
	}{
		{0, false, true, 4}, {1, true, false, 7}, {1, false, true, 3}, {2, true, false, 6},
		{3, true, false, 5}, {3, false, true, 1}, {4, false, true, 0}, {5, true, false, 3},
		{6, true, false, 2}, {7, true, false, 1}, {7, false, true, 5},
	}
	for _, test := range tests {
		if got := Reflect(test.direction, test.blockX, test.blockY); got != test.want {
			t.Errorf("Reflect(%d,%t,%t)=%d want %d", test.direction, test.blockX, test.blockY, got, test.want)
		}
	}
}

// TestRebounds verifies deterministic cardinal and diagonal alternatives.
func TestRebounds(t *testing.T) {
	tests := []struct {
		direction uint8
		want      [3]uint8
	}{
		{0, [3]uint8{4, 4, 4}},
		{1, [3]uint8{7, 3, 5}},
		{2, [3]uint8{6, 6, 6}},
		{3, [3]uint8{5, 1, 7}},
		{4, [3]uint8{0, 0, 0}},
		{5, [3]uint8{3, 7, 1}},
		{6, [3]uint8{2, 2, 2}},
		{7, [3]uint8{1, 5, 3}},
	}
	for _, test := range tests {
		if got := Rebounds(test.direction); got != test.want {
			t.Errorf("Rebounds(%d)=%v want %v", test.direction, got, test.want)
		}
	}
}

// TestGoalScores verifies directional scoring.
func TestGoalScores(t *testing.T) {
	for rotation := uint8(0); rotation < 8; rotation++ {
		front := (rotation + 4) % 8
		for _, direction := range []uint8{(front + 7) % 8, front, (front + 1) % 8} {
			if !GoalScores(direction, rotation) {
				t.Errorf("front entry %d rejected at %d", direction, rotation)
			}
		}
		if GoalScores(rotation, rotation) {
			t.Errorf("back entry accepted at %d", rotation)
		}
		if GoalScores((front+2)%8, rotation) || GoalScores((front+6)%8, rotation) {
			t.Errorf("side entry accepted at %d", rotation)
		}
	}
}

// TestScoreboard verifies increment, decrement, reset, and wrap.
func TestScoreboard(t *testing.T) {
	board := Scoreboard{Value: 99}
	if board.Add(1) != 0 || board.Add(-1) != 99 {
		t.Fatalf("unexpected wrap: %+v", board)
	}
	board.Reset()
	if board.Value != 0 {
		t.Fatal("reset failed")
	}
}

// BenchmarkPhysics measures rebound and goal-face calculations on the tick path.
func BenchmarkPhysics(b *testing.B) {
	b.ReportAllocs()
	accepted := 0
	for index := 0; index < b.N; index++ {
		direction := uint8(index % 8)
		accepted += int(Rebounds(direction)[0])
		if GoalScores(direction, uint8((index+4)%8)) {
			accepted++
		}
	}
	if accepted < 0 {
		b.Fatal(accepted)
	}
}
