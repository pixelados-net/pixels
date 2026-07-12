package essential

import (
	"context"
	"testing"
	"time"

	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
)

// stateRecorder records final durable state changes.
type stateRecorder struct {
	// params stores recorded state changes.
	params []furnitureservice.StateParams
}

// TestServiceRoutesSupportedInteractions verifies central dispatch and validation.
func TestServiceRoutesSupportedInteractions(t *testing.T) {
	service := New(nil, nil, nil, nil, nil, bus.New(), nil, nil)
	service.SetSource(fixedSource(0))
	if value := (defaultSource{}).IntN(2); value < 0 || value >= 2 {
		t.Fatalf("default source returned %d", value)
	}
	if handled, err := service.Use(context.Background(), Request{}); handled || err != nil {
		t.Fatalf("invalid request handled=%t err=%v", handled, err)
	}
	item := essentialItem("pressureplate", 2)
	active := essentialRoom(t, item, 1)
	if handled, err := service.Use(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); !handled || err != nil {
		t.Fatalf("pressure plate dispatch handled=%t err=%v", handled, err)
	}
	for _, kind := range []string{"dice", "onewaygate", "handitem", "cannon"} {
		item.Definition.InteractionType = kind
		item.Point = grid.MustPoint(5, 0)
		if handled, err := service.Use(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); !handled || err != nil {
			t.Fatalf("%s dispatch handled=%t err=%v", kind, handled, err)
		}
	}
	item.Definition.InteractionType = "unsupported"
	if handled, err := service.Use(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); handled || err != nil {
		t.Fatalf("unsupported dispatch handled=%t err=%v", handled, err)
	}
}

// TestMovementSubscriptionsProjectPlateState verifies bus registration and walk routing.
func TestMovementSubscriptionsProjectPlateState(t *testing.T) {
	item := essentialItem("colorplate", 3)
	registry := roomlive.NewRegistry(nil)
	active, err := registry.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	roomGrid, err := grid.Parse("000000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse active grid: %v", err)
	}
	if err := active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{item}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatalf("load active world: %v", err)
	}
	if _, err := active.Join(roomlive.Occupant{PlayerID: 1, Username: "test", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	local := bus.New()
	service := &Service{runtime: registry, random: fixedSource(0)}
	lifecycle := fxtest.NewLifecycle(t)
	if err := Register(lifecycle, local, service); err != nil {
		t.Fatalf("register interactions: %v", err)
	}
	if err := lifecycle.Start(context.Background()); err != nil {
		t.Fatalf("start lifecycle: %v", err)
	}
	on := furniturewalkedon.Payload{RoomID: 9, ItemID: item.ID, PlayerID: 1}
	off := furniturewalkedoff.Payload{RoomID: 9, ItemID: item.ID, PlayerID: 1}
	_ = local.Publish(context.Background(), bus.Event{Name: furniturewalkedon.Name, Payload: "invalid"})
	_ = local.Publish(context.Background(), bus.Event{Name: furniturewalkedoff.Name, Payload: "invalid"})
	missing := furniturewalkedon.Payload{RoomID: 9, ItemID: 999, PlayerID: 1}
	_ = local.Publish(context.Background(), bus.Event{Name: furniturewalkedon.Name, Payload: missing})
	_ = local.Publish(context.Background(), bus.Event{Name: furniturewalkedon.Name, Payload: on})
	_ = local.Publish(context.Background(), bus.Event{Name: furniturewalkedoff.Name, Payload: off})
	updated, _ := active.FurnitureItem(item.ID)
	if updated.ExtraData != "0" {
		t.Fatalf("expected balanced plate state, got %q", updated.ExtraData)
	}
	pressure := essentialItem("pressureplate", 2)
	pressure.Definition.AllowWalk = true
	pressure.Definition.StackHeight = 0
	if _, err := active.ReloadFurniture(item.ID, &pressure); err != nil {
		t.Fatalf("reload pressure plate: %v", err)
	}
	_ = local.Publish(context.Background(), bus.Event{Name: furniturewalkedon.Name, Payload: on})
	_ = local.Publish(context.Background(), bus.Event{Name: furniturewalkedoff.Name, Payload: off})
	active.RunScheduled(time.Now().Add(time.Second))
	handItem := essentialItem("handitem_tile", 2)
	handItem.Definition.AllowWalk = true
	handItem.Definition.StackHeight = 0
	handItem.Definition.CustomParams = "13"
	if _, err := active.ReloadFurniture(item.ID, &handItem); err != nil {
		t.Fatalf("reload hand item tile: %v", err)
	}
	_ = local.Publish(context.Background(), bus.Event{Name: furniturewalkedon.Name, Payload: on})
	unit, _ := active.Unit(1)
	if unit.HandItem != 13 {
		t.Fatalf("expected hand item tile delivery, got %d", unit.HandItem)
	}
	if err := lifecycle.Stop(context.Background()); err != nil {
		t.Fatalf("stop lifecycle: %v", err)
	}
	_, _, _ = registry.Close(context.Background(), 9)
}

// UpdateState records one state mutation.
func (recorder *stateRecorder) UpdateState(_ context.Context, params furnitureservice.StateParams) (furnituremodel.Item, error) {
	recorder.params = append(recorder.params, params)

	return furnituremodel.Item{}, nil
}

// fixedSource returns one deterministic bounded value.
type fixedSource int

// IntN returns the configured value bounded by limit.
func (source fixedSource) IntN(limit int) int {
	if limit <= 0 {
		return 0
	}

	return int(source) % limit
}

// essentialRoom creates a loaded room with one furniture item and players.
func essentialRoom(t testing.TB, item worldfurniture.Item, playerIDs ...int64) *roomlive.Room {
	t.Helper()
	roomGrid, err := grid.Parse("000000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 25})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if err := active.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid, Furniture: []worldfurniture.Item{item},
		Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}
	for _, playerID := range playerIDs {
		_, err := active.Join(roomlive.Occupant{PlayerID: playerID, Username: "test", ConnectionID: "conn", ConnectionKind: "websocket"})
		if err != nil {
			t.Fatalf("join player %d: %v", playerID, err)
		}
	}

	return active
}

// essentialItem creates one runtime item at x=1.
func essentialItem(interactionType string, modes int) worldfurniture.Item {
	return worldfurniture.Item{
		ID: 10, Point: grid.MustPoint(1, 0), Rotation: worldunit.RotationEast, ExtraData: "0",
		Definition: worldfurniture.Definition{
			InteractionType: interactionType, InteractionModesCount: modes,
			Width: 1, Length: 1, StackHeight: 1,
		},
	}
}
