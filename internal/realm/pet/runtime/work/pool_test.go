package work

import (
	"sync/atomic"
	"testing"
)

// TestPoolDrainsAcceptedWork verifies fixed workers complete accepted jobs.
func TestPoolDrainsAcceptedWork(t *testing.T) {
	pool := New(8, 2, nil)
	pool.Start()
	count := atomic.Int32{}
	for range 8 {
		if !pool.Dispatch(func() { count.Add(1) }) {
			t.Fatal("expected bounded job acceptance")
		}
	}
	pool.Stop()
	if count.Load() != 8 {
		t.Fatalf("expected 8 completed jobs, got %d", count.Load())
	}
	if pool.Dispatch(func() {}) {
		t.Fatal("expected stopped pool rejection")
	}
}
