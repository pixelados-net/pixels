package effect

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestExpireProjectsConsumedAndRemainingCharges verifies aggregate expiry branches.
func TestExpireProjectsConsumedAndRemainingCharges(t *testing.T) {
	store := newMemoryStore()
	store.expired = []Expiration{
		{PlayerID: 7, EffectID: 101, Selected: true},
		{PlayerID: 8, EffectID: 102, RemainingCharges: 2},
	}
	service := New(store, nil, nil, nil, nil, bus.New())
	if err := service.expire(context.Background(), time.Unix(100, 0)); err != nil {
		t.Fatal(err)
	}
}

// TestEffectSchedulerLifecycle verifies the single global loop is cancellable.
func TestEffectSchedulerLifecycle(t *testing.T) {
	lifecycle := fxtest.NewLifecycle(t)
	service := New(newMemoryStore(), nil, nil, nil, nil, nil)
	RegisterScheduler(lifecycle, service, zap.NewNop())
	lifecycle.RequireStart().RequireStop()
}
