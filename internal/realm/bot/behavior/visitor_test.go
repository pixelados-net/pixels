package behavior

import (
	"context"
	"strings"
	"testing"
	"time"

	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// TestVisitorLogTargetsOwnerAndFormatsVisits verifies summary and detail behavior.
func TestVisitorLogTargetsOwnerAndFormatsVisits(t *testing.T) {
	actions := &actionRecorder{visits: []sdkbot.Visit{{PlayerName: "Alice", EnteredAt: time.Date(2026, 7, 14, 10, 30, 0, 0, time.UTC)}}}
	behavior := VisitorLog{}
	bot := sdkbot.Bot{OwnerPlayerID: 1}
	if err := behavior.OnUserEnter(context.Background(), bot, 2, actions); err != nil || len(actions.talks) != 0 {
		t.Fatalf("non-owner talks=%v err=%v", actions.talks, err)
	}
	if err := behavior.OnUserEnter(context.Background(), bot, 1, actions); err != nil || len(actions.talks) != 1 || actions.talks[0] != "bots.visitor.visits:1" {
		t.Fatalf("summary=%v err=%v", actions.talks, err)
	}
	if err := behavior.OnUserSay(context.Background(), bot, sdkbot.Message{PlayerID: 1, Text: "sí"}, actions); err != nil || len(actions.talks) != 2 || !strings.Contains(actions.talks[1], "Alice") {
		t.Fatalf("details=%v err=%v", actions.talks, err)
	}
}
