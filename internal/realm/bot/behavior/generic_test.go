package behavior

import (
	"context"
	"testing"

	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// actionRecorder records behavior-facing decisions.
type actionRecorder struct {
	// walks counts random movement requests.
	walks int
	// talks stores delivered messages.
	talks []string
	// visits stores visitor-log source rows.
	visits []sdkbot.Visit
}

// RandomWalk records one movement decision.
func (actions *actionRecorder) RandomWalk(context.Context, sdkbot.Bot) error {
	actions.walks++
	return nil
}

// Talk records one delivered message.
func (actions *actionRecorder) Talk(_ context.Context, _ sdkbot.Bot, message string, _ sdkbot.Scope, _ int64) error {
	actions.talks = append(actions.talks, message)
	return nil
}

// ServeKeyword accepts bartender calls without side effects.
func (*actionRecorder) ServeKeyword(context.Context, sdkbot.Bot, sdkbot.Message) (bool, error) {
	return true, nil
}

// Visits returns configured visitor rows.
func (actions *actionRecorder) Visits(context.Context, sdkbot.Bot, int64) ([]sdkbot.Visit, error) {
	return actions.visits, nil
}

// TestGenericCycleSuppressesWalkWhileFollowing verifies follow precedence.
func TestGenericCycleSuppressesWalkWhileFollowing(t *testing.T) {
	actions := &actionRecorder{}
	behavior := Generic{}
	bot := sdkbot.Bot{CanWalk: true, WalkDue: true, FollowingPlayerID: 5, ChatDue: true, ChatMessage: "hello"}
	if err := behavior.OnCycle(context.Background(), bot, actions); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	if actions.walks != 0 || len(actions.talks) != 1 || actions.talks[0] != "hello" {
		t.Fatalf("walks=%d talks=%v", actions.walks, actions.talks)
	}
	bot.FollowingPlayerID = 0
	if err := behavior.OnCycle(context.Background(), bot, actions); err != nil || actions.walks != 1 {
		t.Fatalf("walks=%d err=%v", actions.walks, err)
	}
}
