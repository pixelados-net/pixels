package priority

import "testing"

// TestPriorityOrder verifies larger callbacks run first and monitors run last.
func TestPriorityOrder(t *testing.T) {
	if !(Highest > High && High > Normal && Normal > Low && Low > Lowest && Lowest > Monitor) {
		t.Fatalf("unexpected callback priority order")
	}
}
