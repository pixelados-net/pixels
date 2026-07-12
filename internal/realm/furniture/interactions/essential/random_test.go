package essential

import (
	"context"
	"testing"
	"time"
)

// TestDiceRollsOnceAndSettles verifies sentinel, guard, and deterministic result.
func TestDiceRollsOnceAndSettles(t *testing.T) {
	item := essentialItem("dice", 6)
	active := essentialRoom(t, item, 1)
	states := &stateRecorder{}
	service := &Service{states: states, random: fixedSource(2)}
	request := Request{PlayerID: 1, Room: active, Item: item}
	if err := service.useRandom(context.Background(), request); err != nil {
		t.Fatalf("start roll: %v", err)
	}
	rolling, _ := active.FurnitureItem(item.ID)
	if rolling.ExtraData != "-1" {
		t.Fatalf("expected rolling sentinel, got %q", rolling.ExtraData)
	}
	request.Item = rolling
	if err := service.useRandom(context.Background(), request); err != nil {
		t.Fatalf("repeat roll: %v", err)
	}
	active.RunScheduled(time.Now().Add(2 * time.Second))
	settled, _ := active.FurnitureItem(item.ID)
	if settled.ExtraData != "3" || len(states.params) != 1 {
		t.Fatalf("unexpected settled state=%q writes=%d", settled.ExtraData, len(states.params))
	}
}

// TestParseRandomParams verifies Arcturus-compatible definition parameters.
func TestParseRandomParams(t *testing.T) {
	states, delay := parseRandomParams("states=5,delay=750")
	if states != 5 || delay != 750*time.Millisecond {
		t.Fatalf("unexpected policy states=%d delay=%s", states, delay)
	}
}

// TestRandomFamiliesApplyTheirPolicies verifies wheel and configurable random state delays.
func TestRandomFamiliesApplyTheirPolicies(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		params   string
		modes    int
		expected string
	}{
		{name: "color wheel", kind: "colorwheel", modes: 4, expected: "2"},
		{name: "random state", kind: "random_state", params: "states=3,delay=25", modes: 1, expected: "2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := essentialItem(test.kind, test.modes)
			item.Definition.CustomParams = test.params
			active := essentialRoom(t, item, 1)
			states := &stateRecorder{}
			service := &Service{states: states, random: fixedSource(1)}
			handled, err := service.Use(context.Background(), Request{PlayerID: 1, Room: active, Item: item})
			if err != nil || !handled {
				t.Fatalf("use random interaction handled=%t err=%v", handled, err)
			}
			active.RunScheduled(time.Now().Add(4 * time.Second))
			updated, _ := active.FurnitureItem(item.ID)
			if updated.ExtraData != test.expected || len(states.params) != 1 {
				t.Fatalf("unexpected state=%q writes=%d", updated.ExtraData, len(states.params))
			}
		})
	}
}

// TestRandomHelpersHandleBoundaries verifies geometry and malformed policy fallbacks.
func TestRandomHelpersHandleBoundaries(t *testing.T) {
	states, delay := parseRandomParams("states=nope,delay=-1")
	if states != 0 || delay != 0 {
		t.Fatalf("unexpected fallback states=%d delay=%s", states, delay)
	}
	item := essentialItem("switch", 2)
	points := activatorPoints(item)
	if len(points) == 0 || distance(points[0], points[0]) != 0 {
		t.Fatalf("unexpected activators %#v", points)
	}
}

// TestCloseDiceResetsOnlySettledValues verifies Nitro's dedicated close behavior.
func TestCloseDiceResetsOnlySettledValues(t *testing.T) {
	item := essentialItem("dice", 6)
	item.ExtraData = "4"
	active := essentialRoom(t, item, 1)
	states := &stateRecorder{}
	service := &Service{states: states, random: fixedSource(0)}
	request := Request{PlayerID: 1, Room: active, Item: item}
	if err := service.CloseDice(context.Background(), request); err != nil {
		t.Fatalf("close dice: %v", err)
	}
	closed, _ := active.FurnitureItem(item.ID)
	if closed.ExtraData != "0" || len(states.params) != 1 {
		t.Fatalf("unexpected closed state=%q writes=%d", closed.ExtraData, len(states.params))
	}
	request.Item.ExtraData = randomRollingState
	if err := service.CloseDice(context.Background(), request); err != nil || len(states.params) != 1 {
		t.Fatalf("rolling dice close err=%v writes=%d", err, len(states.params))
	}
}

// TestRandomStateWithoutPolicyIsIgnored verifies incomplete definitions stay inert.
func TestRandomStateWithoutPolicyIsIgnored(t *testing.T) {
	item := essentialItem("random_state", 1)
	active := essentialRoom(t, item, 1)
	service := &Service{random: fixedSource(0)}
	if err := service.useRandom(context.Background(), Request{PlayerID: 1, Room: active, Item: item}); err != nil {
		t.Fatalf("ignore empty random policy: %v", err)
	}
	updated, _ := active.FurnitureItem(item.ID)
	if updated.ExtraData != "0" {
		t.Fatalf("unexpected random state %q", updated.ExtraData)
	}
}
