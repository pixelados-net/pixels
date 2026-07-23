package tick

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestTargetFuncHandlesTick verifies function targets implement Target.
func TestTargetFuncHandlesTick(t *testing.T) {
	expected := Tick{At: time.Unix(10, 0), Delta: time.Second, Sequence: 4}
	var handled Tick
	target := TargetFunc(func(ctx context.Context, tick Tick) error {
		handled = tick
		return nil
	})

	if err := target.Tick(context.Background(), expected); err != nil {
		t.Fatalf("handle tick: %v", err)
	}
	if handled != expected {
		t.Fatalf("expected %#v, got %#v", expected, handled)
	}
}

// TestTargetFuncReturnsError verifies target errors are preserved.
func TestTargetFuncReturnsError(t *testing.T) {
	expected := errors.New("tick failed")
	target := TargetFunc(func(context.Context, Tick) error {
		return expected
	})

	err := target.Tick(context.Background(), Tick{})
	if !errors.Is(err, expected) {
		t.Fatalf("expected target error, got %v", err)
	}
}

// TestConfigValidate verifies ticker interval validation.
func TestConfigValidate(t *testing.T) {
	if err := (Config{Interval: time.Second}).Validate(); err != nil {
		t.Fatalf("validate config: %v", err)
	}
	if err := (Config{}).Validate(); !errors.Is(err, ErrInvalidInterval) {
		t.Fatalf("expected invalid interval error, got %v", err)
	}
}
