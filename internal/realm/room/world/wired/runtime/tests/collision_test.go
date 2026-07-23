package tests

import (
	"context"
	"testing"
	"time"

	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// collisionFurniture returns one derived collision for each furniture movement execution.
type collisionFurniture struct {
	// calls stores the number of independent movement attempts.
	calls int
}

// ExecuteFurniture records a movement and derives its collided actor context.
func (service *collisionFurniture) ExecuteFurniture(_ context.Context, _ effect.FurnitureOperation, _ *configuration.Node, event trigger.Event) (effect.Result, error) {
	service.calls++
	return effect.Result{Status: effect.Applied, Derived: []trigger.Event{{
		Kind: trigger.Collision, RoomID: event.RoomID, ActorKind: event.ActorKind,
		ActorID: event.ActorID, PlayerID: event.PlayerID, SourceItem: 10,
	}}}, nil
}

// TestDerivedCollisionExecutesAgainOnANewEvent verifies trace-local loop prevention never consumes future collisions.
func TestDerivedCollisionExecutesAgainOnANewEvent(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_says_something", X: 1, Y: 1, StringParam: "collide"},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_flee", X: 1, Y: 1, SelectionMode: 1, Targets: []record.Target{{ItemID: 10}}},
		{ItemID: 3, RoomID: 1, Interaction: "wf_trg_collision", X: 2, Y: 2},
		{ItemID: 4, RoomID: 1, Interaction: "wf_act_show_message", X: 2, Y: 2, StringParam: "PASS collision fired."},
	}
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	avatarService, furnitureService := &avatar{}, &collisionFurniture{}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	executor := effect.New(effect.Services{Avatar: avatarService, Furniture: furnitureService})
	engine := wiredruntime.New(roomwired.Config{Enabled: true}, store{records: records}, compiler, executor, nil, nil, nil)
	now := time.Now()
	if err = engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		event := trigger.Event{Kind: trigger.Say, RoomID: 1, ActorKind: trigger.ActorPlayer, ActorID: 7, PlayerID: 7, Message: "collide"}
		if _, err = engine.Process(context.Background(), event, now.Add(time.Duration(attempt)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	if furnitureService.calls != 2 || avatarService.calls != 2 {
		t.Fatalf("furniture calls=%d message calls=%d", furnitureService.calls, avatarService.calls)
	}
}
