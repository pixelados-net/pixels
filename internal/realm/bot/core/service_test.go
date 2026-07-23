package core

import (
	"sync/atomic"
	"testing"

	botpolicy "github.com/niflaot/pixels/internal/realm/bot/policy"
)

// TestStopDrainsAcceptedJobs verifies shutdown preserves already queued persistence work.
func TestStopDrainsAcceptedJobs(t *testing.T) {
	service := New(botpolicy.Config{}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	service.Start()
	release := make(chan struct{})
	var completed atomic.Int64
	for range 4 {
		service.dispatch(func() {
			<-release
			completed.Add(1)
		})
	}
	for range 32 {
		service.dispatch(func() { completed.Add(1) })
	}
	done := make(chan struct{})
	go func() {
		service.Stop()
		close(done)
	}()
	<-service.stopped
	close(release)
	<-done
	if completed.Load() != 36 {
		t.Fatalf("completed=%d, want 36", completed.Load())
	}
}
