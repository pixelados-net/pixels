package settings

import (
	"context"
	"testing"
	"time"

	"go.uber.org/fx"
)

// settingsLifecycle captures one settings writer lifecycle hook.
type settingsLifecycle struct {
	hook fx.Hook
}

// Append captures one lifecycle hook.
func (lifecycle *settingsLifecycle) Append(hook fx.Hook) { lifecycle.hook = hook }

// TestWriterCoalescesFields verifies latest values persist once per field.
func TestWriterCoalescesFields(t *testing.T) {
	store := &memoryStore{}
	writer := NewWriter(New(store), nil, Config{FlushInterval: time.Second, PendingLimit: 1})
	if !writer.EnqueueVolume(1, 10, 20, 30) || !writer.EnqueueVolume(1, 40, 50, 60) || !writer.EnqueueOldChat(1, true) || !writer.EnqueueCameraFollowBlocked(1, true) {
		t.Fatal("expected same-player settings to coalesce")
	}
	if writer.EnqueueOldChat(2, true) {
		t.Fatal("expected distinct-player pending limit")
	}
	writer.flush(context.Background())
	if store.record.VolumeSystem != 40 || store.record.VolumeFurniture != 50 || store.record.VolumeTrax != 60 || !store.record.OldChat || !store.record.CameraFollowBlocked {
		t.Fatalf("unexpected persisted settings %#v", store.record)
	}
}

// TestWriterLifecycleFlushesOnStop verifies bounded startup and shutdown ownership.
func TestWriterLifecycleFlushesOnStop(t *testing.T) {
	store := &memoryStore{}
	writer := NewWriter(New(store), nil, Config{})
	if !writer.EnqueueOldChat(1, true) {
		t.Fatal("enqueue rejected")
	}
	lifecycle := &settingsLifecycle{}
	RegisterWriter(lifecycle, writer)
	if err := lifecycle.hook.OnStart(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := lifecycle.hook.OnStop(ctx); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if !store.record.OldChat {
		t.Fatalf("record=%#v", store.record)
	}
}

// BenchmarkWriterEnqueueVolume measures the warmed coalescing path.
func BenchmarkWriterEnqueueVolume(benchmark *testing.B) {
	writer := NewWriter(New(&memoryStore{}), nil, Config{FlushInterval: time.Second, PendingLimit: 1})
	writer.EnqueueVolume(1, 10, 20, 30)
	benchmark.ReportAllocs()
	for range benchmark.N {
		if !writer.EnqueueVolume(1, 10, 20, 30) {
			benchmark.Fatal("enqueue rejected")
		}
	}
}
