package furniture

import (
	"context"
	"testing"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// collisionRuntimeStore supplies one immutable end-to-end collision generation.
type collisionRuntimeStore struct {
	// records stores configured WIRED nodes.
	records []record.Config
}

// LoadRoom returns the configured collision generation.
func (store collisionRuntimeStore) LoadRoom(context.Context, int64) ([]record.Config, error) {
	return append([]record.Config(nil), store.records...), nil
}

// Find is unused by the collision runtime test.
func (collisionRuntimeStore) Find(context.Context, int64, int64) (record.Config, bool, error) {
	return record.Config{}, false, nil
}

// Save is unused by the collision runtime test.
func (collisionRuntimeStore) Save(context.Context, record.Config, int64) (record.Config, error) {
	return record.Config{}, nil
}

// SaveRewardConfig is unused by the collision runtime test.
func (collisionRuntimeStore) SaveRewardConfig(context.Context, record.Config, int64, []record.Reward) (record.Config, error) {
	return record.Config{}, nil
}

// CleanupItem is unused by the collision runtime test.
func (collisionRuntimeStore) CleanupItem(context.Context, int64) error { return nil }

// Capture is unused by the collision runtime test.
func (collisionRuntimeStore) Capture(context.Context, int64, int64) ([]record.Target, error) {
	return nil, nil
}

// collisionRuntimeAvatar records observable collision messages.
type collisionRuntimeAvatar struct {
	// calls stores message executions.
	calls int
	// actorID stores the last collided actor.
	actorID int64
}

// ExecuteAvatar records one player-facing collision effect.
func (avatar *collisionRuntimeAvatar) ExecuteAvatar(_ context.Context, _ effect.AvatarOperation, _ *configuration.Node, event trigger.Event) (effect.Result, error) {
	avatar.calls++
	avatar.actorID = event.ActorID
	return effect.Result{Status: effect.Applied}, nil
}

// TestMovementCollisionExecutesObservableStackTwice verifies the real movement service and engine together.
func TestMovementCollisionExecutesObservableStackTwice(t *testing.T) {
	rooms, manager, active := furnitureRoom(t)
	if _, err := active.Join(occupantForCollisionTest()); err != nil {
		t.Fatal(err)
	}
	records := []record.Config{
		{ItemID: 1, RoomID: active.ID(), Interaction: "wf_trg_says_something", X: 1, Y: 1, StringParam: "collision"},
		{ItemID: 2, RoomID: active.ID(), Interaction: "wf_act_move_to_dir", X: 1, Y: 1, IntParams: []int32{6}, SelectionMode: 1, Targets: []record.Target{{ItemID: 10}}},
		{ItemID: 3, RoomID: active.ID(), Interaction: "wf_trg_collision", X: 2, Y: 2},
		{ItemID: 4, RoomID: active.ID(), Interaction: "wf_act_show_message", X: 2, Y: 2, StringParam: "PASS collision fired."},
	}
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	avatar := &collisionRuntimeAvatar{}
	executor := effect.New(effect.Services{Furniture: New(rooms, manager, nil), Avatar: avatar})
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	engine := wiredruntime.New(roomwired.Config{Enabled: true}, collisionRuntimeStore{records: records}, compiler, executor, nil, nil, nil)
	if err = engine.Reload(context.Background(), active.ID(), time.Now()); err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		event := trigger.Event{Kind: trigger.Say, RoomID: active.ID(), ActorKind: trigger.ActorPlayer, ActorID: 7, PlayerID: 7, Message: "collision"}
		if _, err = engine.Process(context.Background(), event, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	if avatar.calls != 2 || avatar.actorID != 7 {
		t.Fatalf("message calls=%d actor=%d", avatar.calls, avatar.actorID)
	}
	item, _ := active.FurnitureItem(10)
	if item.Point.X != 1 || item.Point.Y != 0 {
		t.Fatalf("collided furniture moved to %+v", item.Point)
	}
}

// occupantForCollisionTest creates the player blocking the furniture destination.
func occupantForCollisionTest() roomlive.Occupant {
	return roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}
}
