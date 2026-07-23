package runtime

import (
	"context"
	"errors"
	"testing"
	"time"
)

// schedulerCycle records subscription reconciliation passes.
type schedulerCycle struct {
	// calls receives one signal for every reconciliation pass.
	calls chan struct{}
	// err stores the configured reconciliation failure.
	err error
}

// RunCycle records one reconciliation pass.
func (cycle *schedulerCycle) RunCycle(context.Context) error {
	cycle.calls <- struct{}{}
	return cycle.err
}

// TestSchedulerReconcilesBeforeReturningFromStart verifies startup state is immediately current.
func TestSchedulerReconcilesBeforeReturningFromStart(t *testing.T) {
	cycle := &schedulerCycle{calls: make(chan struct{}, 2)}
	scheduler := &Scheduler{interval: time.Hour, service: cycle}
	if err := scheduler.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = scheduler.Stop(context.Background()) })
	select {
	case <-cycle.calls:
	default:
		t.Fatal("startup reconciliation did not finish before Start returned")
	}
}

// TestSchedulerStartPropagatesReconciliationFailure verifies startup does not hide database failures.
func TestSchedulerStartPropagatesReconciliationFailure(t *testing.T) {
	expected := errors.New("reconciliation failed")
	cycle := &schedulerCycle{calls: make(chan struct{}, 1), err: expected}
	scheduler := &Scheduler{interval: time.Hour, service: cycle}
	if err := scheduler.Start(context.Background()); !errors.Is(err, expected) {
		t.Fatalf("error=%v want=%v", err, expected)
	}
	if scheduler.cancel != nil {
		t.Fatal("scheduler loop started after reconciliation failure")
	}
}
