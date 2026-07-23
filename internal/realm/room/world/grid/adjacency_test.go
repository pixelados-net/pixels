package grid

import "testing"

// TestAdjacent verifies edge, corner, equal, and distant point classification.
func TestAdjacent(t *testing.T) {
	cases := []struct {
		name  string
		left  Point
		right Point
		want  bool
	}{
		{name: "edge", left: MustPoint(1, 1), right: MustPoint(2, 1), want: true},
		{name: "corner", left: MustPoint(1, 1), right: MustPoint(2, 2), want: true},
		{name: "equal", left: MustPoint(1, 1), right: MustPoint(1, 1), want: false},
		{name: "distant", left: MustPoint(1, 1), right: MustPoint(3, 1), want: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := Adjacent(test.left, test.right); got != test.want {
				t.Fatalf("Adjacent()=%v want=%v", got, test.want)
			}
		})
	}
}
