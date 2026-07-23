package bot

import (
	"context"
	"errors"
	"testing"
	"time"

	botbehavior "github.com/niflaot/pixels/internal/realm/bot/behavior"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botpolicy "github.com/niflaot/pixels/internal/realm/bot/policy"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"go.uber.org/zap"
)

// botStore supplies one placed bot to the core runtime.
type botStore struct {
	botrecord.Store
	// bots stores placed test records.
	bots []botrecord.Bot
}

// Room returns placed bot records.
func (store *botStore) Room(context.Context, int64) ([]botrecord.Bot, error) { return store.bots, nil }

// TestExecuteBotFailsClosedWithoutRuntime verifies an unavailable bot realm cannot apply effects.
func TestExecuteBotFailsClosedWithoutRuntime(t *testing.T) {
	service := New(roomlive.NewRegistry(nil), nil)
	result, err := service.ExecuteBot(context.Background(), effect.BotTalk, &configuration.Node{}, trigger.Event{RoomID: 1})
	if err != nil || result.Status != effect.Blocked {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

// TestAppliedMapsOperationErrors verifies bot adapter status mapping.
func TestAppliedMapsOperationErrors(t *testing.T) {
	result, err := applied(nil)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("successful result=%+v err=%v", result, err)
	}
	want := errors.New("bot failure")
	result, err = applied(want)
	if !errors.Is(err, want) || result.Status != effect.Blocked {
		t.Fatalf("blocked result=%+v err=%v", result, err)
	}
}

// TestExecuteBotUsesExistingRuntime verifies movement, speech, follow, and hand-item effects reuse bot core.
func TestExecuteBotUsesExistingRuntime(t *testing.T) {
	rooms, active, bots := botRoom(t)
	service := New(rooms, bots)
	event := trigger.Event{RoomID: active.ID(), PlayerID: 7, ActorID: 7}
	move := &configuration.Node{Parameters: configuration.Parameters{Name: "Helper"}, Targets: []record.Target{{ItemID: 11}}}
	result, err := service.ExecuteBot(context.Background(), effect.BotMove, move, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("move result=%+v err=%v", result, err)
	}
	teleport := &configuration.Node{Parameters: configuration.Parameters{Name: "Helper"}, Targets: []record.Target{{ItemID: 10}}}
	result, err = service.ExecuteBot(context.Background(), effect.BotTeleport, teleport, event)
	if err != nil || result.Status != effect.Applied || len(result.Derived) != 1 || result.Derived[0].Kind != trigger.BotReachedFurniture {
		t.Fatalf("teleport result=%+v err=%v", result, err)
	}
	talk := &configuration.Node{Parameters: configuration.Parameters{Name: "Helper", Message: "hello"}}
	result, err = service.ExecuteBot(context.Background(), effect.BotTalk, talk, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("talk result=%+v err=%v", result, err)
	}
	result, err = service.ExecuteBot(context.Background(), effect.BotTalkToAvatar, talk, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("direct talk result=%+v err=%v", result, err)
	}
	result, err = service.ExecuteBot(context.Background(), effect.BotFollowAvatar, talk, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("follow result=%+v err=%v", result, err)
	}
	hand := &configuration.Node{Parameters: configuration.Parameters{Name: "Helper", Values: []int32{5}}}
	result, err = service.ExecuteBot(context.Background(), effect.BotGiveHanditem, hand, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("hand item result=%+v err=%v", result, err)
	}
	unit, _ := active.Unit(7)
	if unit.HandItem != 5 {
		t.Fatalf("hand item=%d", unit.HandItem)
	}
}

// TestExecuteBotSkipsMissingTargetsAndActors verifies optional bot inputs fail closed.
func TestExecuteBotSkipsMissingTargetsAndActors(t *testing.T) {
	rooms, active, bots := botRoom(t)
	service := New(rooms, bots)
	missing := &configuration.Node{Parameters: configuration.Parameters{Name: "Missing"}}
	result, err := service.ExecuteBot(context.Background(), effect.BotTalk, missing, trigger.Event{RoomID: active.ID()})
	if err != nil || result.Status != effect.Skipped {
		t.Fatalf("missing bot result=%+v err=%v", result, err)
	}
	node := &configuration.Node{Parameters: configuration.Parameters{Name: "Helper"}}
	for _, operation := range []effect.BotOperation{effect.BotTalkToAvatar, effect.BotFollowAvatar, effect.BotGiveHanditem} {
		result, err = service.ExecuteBot(context.Background(), operation, node, trigger.Event{RoomID: active.ID()})
		if err != nil || result.Status != effect.Skipped {
			t.Errorf("operation=%d result=%+v err=%v", operation, result, err)
		}
	}
	result, _ = service.ExecuteBot(context.Background(), effect.BotClothes, node, trigger.Event{RoomID: active.ID()})
	if result.Status != effect.Blocked {
		t.Fatalf("empty clothes result=%+v", result)
	}
	result, _ = service.ExecuteBot(context.Background(), effect.BotOperation(255), node, trigger.Event{RoomID: active.ID()})
	if result.Status != effect.Blocked {
		t.Fatalf("unknown operation result=%+v", result)
	}
}

// botRoom creates one active player, bot, and two walkable target furnitures.
func botRoom(t *testing.T) (*roomlive.Registry, *roomlive.Room, *botcore.Service) {
	t.Helper()
	rooms := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := rooms.Activate(roomlive.Snapshot{ID: 33, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("0000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	items := []worldfurniture.Item{
		{ID: 11, Point: grid.MustPoint(2, 0), Definition: worldfurniture.Definition{SpriteID: 2, Width: 1, Length: 1, AllowWalk: true}},
		{ID: 10, Point: grid.MustPoint(3, 0), Definition: worldfurniture.Definition{SpriteID: 3, Width: 1, Length: 1, AllowWalk: true}},
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: items, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	if _, err = rooms.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 7, ConnectionID: "test", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	roomID, x, y, z, rotation := active.ID(), 1, 0, 0.0, int16(worldunit.RotationEast)
	placed := botrecord.Bot{ID: 5, OwnerPlayerID: 1, RoomID: &roomID, BehaviorType: botrecord.BehaviorGeneric, Name: "Helper", Figure: "hd-180-1", Gender: "M", X: &x, Y: &y, Z: &z, Rotation: &rotation}
	behaviors := botbehavior.NewRegistry()
	if err = botbehavior.RegisterBuiltins(behaviors); err != nil {
		t.Fatal(err)
	}
	bots := botcore.New(botpolicy.Config{}, &botStore{bots: []botrecord.Bot{placed}}, rooms, nil, playerlive.NewRegistry(), nil, nil, behaviors, nil, nil, nil, nil, nil, zap.NewNop())
	if err = bots.EnsureRoom(context.Background(), active); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = rooms.Close(context.Background(), active.ID()) })
	return rooms, active, bots
}
