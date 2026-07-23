package timer

import (
	"testing"
	"time"
)

// TestTimerLifecycle verifies start, pause, extension, expiry, and reset.
func TestTimerLifecycle(t *testing.T) {
	timer := New([]time.Duration{time.Second, 2 * time.Second})
	if !timer.Start() || timer.Tick(500*time.Millisecond) || timer.Remaining != 500*time.Millisecond {
		t.Fatalf("unexpected start: %+v", timer)
	}
	if timer.Toggle() || timer.Tick(time.Second) {
		t.Fatal("paused timer advanced")
	}
	if timer.Increase() != 1500*time.Millisecond {
		t.Fatalf("unexpected extension: %v", timer.Remaining)
	}
	timer.Toggle()
	if !timer.Tick(2*time.Second) || timer.Started {
		t.Fatal("timer did not expire")
	}
	timer.Reset()
	if timer.Remaining != 2*time.Second {
		t.Fatal("reset lost selected duration")
	}
}

// TestTimerNormalizesStepsAndWrapsSelection verifies malformed definitions are safe.
func TestTimerNormalizesStepsAndWrapsSelection(t *testing.T) {
	timer := New([]time.Duration{-time.Second, time.Second, 0, 3 * time.Second})
	if len(timer.Steps) != 2 || timer.Remaining != time.Second {
		t.Fatalf("timer=%+v", timer)
	}
	if timer.Increase() != 3*time.Second || timer.Increase() != time.Second {
		t.Fatalf("selection did not wrap: %+v", timer)
	}
	defaults := New(nil)
	if len(defaults.Steps) != len(DefaultSteps) || defaults.Remaining != DefaultSteps[0] {
		t.Fatalf("defaults=%+v", defaults)
	}
}
