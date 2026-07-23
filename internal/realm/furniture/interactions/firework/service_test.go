package firework

import (
	"context"
	"errors"
	"testing"
	"time"

	fireworkcharged "github.com/niflaot/pixels/internal/realm/furniture/events/fireworkcharged"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/pkg/bus"
)

// fireworkStateFixture records durable state transitions.
type fireworkStateFixture struct {
	values []furnitureservice.StateParams
	err    error
}

// UpdateState records one guarded transition.
func (states *fireworkStateFixture) UpdateState(_ context.Context, params furnitureservice.StateParams) (furnituremodel.Item, error) {
	states.values = append(states.values, params)
	return furnituremodel.Item{}, states.err
}

// TestUseFurnitureExplodesPublishesAndRecharges verifies the complete scheduled lifecycle.
func TestUseFurnitureExplodesPublishesAndRecharges(t *testing.T) {
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 160, MaxUsers: 2})
	if err != nil {
		t.Fatal(err)
	}
	states := &fireworkStateFixture{}
	events := bus.New()
	published := 0
	_, err = events.Subscribe(fireworkcharged.Name, bus.PriorityNormal, func(context.Context, bus.Event) error { published++; return nil })
	if err != nil {
		t.Fatal(err)
	}
	service := New(Config{DefaultRecharge: time.Second}, states, nil, events)
	request := essential.Request{PlayerID: 1, Room: active, Item: worldfurniture.Item{ID: 970107, ExtraData: ChargedState, Definition: worldfurniture.Definition{InteractionType: "firework", CustomParams: "1"}}}
	handled, err := service.UseFurniture(context.Background(), request)
	if err != nil || !handled || len(states.values) != 1 || states.values[0].Next != ExplodingState || published != 1 {
		t.Fatalf("handled=%t states=%+v published=%d err=%v", handled, states.values, published, err)
	}
	active.RunScheduled(time.Now().Add(2 * time.Second))
	if len(states.values) != 2 || states.values[1].Next != ChargedState {
		t.Fatalf("states=%+v", states.values)
	}
}

// TestUseFurnitureIgnoresOtherTypesAndUnchargedState verifies no phantom progression.
func TestUseFurnitureIgnoresOtherTypesAndUnchargedState(t *testing.T) {
	active, _ := roomlive.NewRoom(roomlive.Snapshot{ID: 160, MaxUsers: 2})
	states := &fireworkStateFixture{}
	service := New(Config{}, states, nil, nil)
	handled, err := service.UseFurniture(context.Background(), essential.Request{PlayerID: 1, Room: active, Item: worldfurniture.Item{ID: 1, Definition: worldfurniture.Definition{InteractionType: "chair"}}})
	if err != nil || handled {
		t.Fatalf("handled=%t err=%v", handled, err)
	}
	handled, err = service.UseFurniture(context.Background(), essential.Request{PlayerID: 1, Room: active, Item: worldfurniture.Item{ID: 2, ExtraData: "0", Definition: worldfurniture.Definition{InteractionType: "firework"}}})
	if err != nil || !handled || len(states.values) != 0 {
		t.Fatalf("handled=%t states=%+v err=%v", handled, states.values, err)
	}
}

// TestUseFurniturePropagatesPersistenceFailure verifies failed explosions do not schedule work.
func TestUseFurniturePropagatesPersistenceFailure(t *testing.T) {
	expected := errors.New("state failed")
	active, _ := roomlive.NewRoom(roomlive.Snapshot{ID: 160, MaxUsers: 2})
	service := New(Config{}, &fireworkStateFixture{err: expected}, nil, nil)
	request := essential.Request{PlayerID: 1, Room: active, Item: worldfurniture.Item{ID: 2, ExtraData: ChargedState, Definition: worldfurniture.Definition{InteractionType: "firework"}}}
	handled, err := service.UseFurniture(context.Background(), request)
	if !handled || !errors.Is(err, expected) {
		t.Fatalf("handled=%t err=%v", handled, err)
	}
}
