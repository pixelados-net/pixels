package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestInvokeCallbackRecoversAndDisables verifies panic isolation state.
func TestInvokeCallbackRecoversAndDisables(t *testing.T) {
	scope := NewScope("broken")
	err := InvokeCallback(context.Background(), time.Second, scope, "test", zap.NewNop(), func(context.Context) error {
		panic("boom")
	})
	if !errors.Is(err, ErrCallbackPanic) || scope.Enabled() {
		t.Fatalf("expected panic and disabled scope, got %v enabled=%v", err, scope.Enabled())
	}
}

// TestInvokeCallbackTimesOut verifies stalled callbacks cannot block the caller.
func TestInvokeCallbackTimesOut(t *testing.T) {
	scope := NewScope("slow")
	err := InvokeCallback(context.Background(), time.Millisecond, scope, "test", zap.NewNop(), func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if !errors.Is(err, ErrCallbackTimeout) {
		t.Fatalf("expected timeout, got %v", err)
	}
}

// BenchmarkInvokeCallback measures the guarded callback boundary.
func BenchmarkInvokeCallback(b *testing.B) {
	scope := NewScope("bench")
	log := zap.NewNop()
	ctx := context.Background()
	callback := func(context.Context) error { return nil }
	b.ReportAllocs()
	for range b.N {
		if err := InvokeCallback(ctx, time.Second, scope, "benchmark", log, callback); err != nil {
			b.Fatal(err)
		}
	}
}
