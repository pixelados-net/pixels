package achievement

import (
	"context"
	"testing"

	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
)

// TestRegisterLifecycleWarmsAndReleasesSnapshots verifies online badge lifecycle ownership.
func TestRegisterLifecycleWarmsAndReleasesSnapshots(t *testing.T) {
	local := bus.New()
	lifecycle := fxtest.NewLifecycle(t)
	service := New(&achievementStore{badges: []Badge{{ID: 1, Code: "ADM", Equipped: true, Slot: 1}}})
	if err := RegisterLifecycle(lifecycle, local, service); err != nil {
		t.Fatal(err)
	}
	lifecycle.RequireStart()
	if err := local.Publish(context.Background(), bus.Event{Name: playerconnected.Name, Payload: "invalid"}); err != nil {
		t.Fatal(err)
	}
	if err := local.Publish(context.Background(), bus.Event{Name: playerconnected.Name, Payload: playerconnected.Payload{PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	if wearing, loaded := service.Wearing(7, "ADM"); !wearing || !loaded {
		t.Fatalf("wearing=%v loaded=%v", wearing, loaded)
	}
	if err := local.Publish(context.Background(), bus.Event{Name: playerdisconnected.Name, Payload: playerdisconnected.Payload{PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	if _, loaded := service.Wearing(7, "ADM"); loaded {
		t.Fatal("disconnected badge snapshot retained")
	}
	lifecycle.RequireStop()
}
