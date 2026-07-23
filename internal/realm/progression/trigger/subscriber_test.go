package trigger

import (
	"context"
	"reflect"
	"testing"

	fireworkcharged "github.com/niflaot/pixels/internal/realm/furniture/events/fireworkcharged"
	playerauthenticated "github.com/niflaot/pixels/internal/realm/player/events/authenticated"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	roomkicked "github.com/niflaot/pixels/internal/realm/room/control/events/kicked"
	roommuted "github.com/niflaot/pixels/internal/realm/room/control/events/muted"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/events/settingsupdated"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/events/wordfiltermodified"
	roomstaffpicked "github.com/niflaot/pixels/internal/realm/room/record/events/staffpicked"
	subscriptionpayday "github.com/niflaot/pixels/internal/realm/subscription/events/payday"
	"github.com/niflaot/pixels/pkg/bus"
)

// progressCall describes one normalized test trigger.
type progressCall struct {
	// playerID identifies the recipient.
	playerID int64
	// key identifies the normalized trigger.
	key string
	// data stores goal metadata.
	data string
	// amount stores the progress delta.
	amount int64
	// daily reports UTC-day idempotence.
	daily bool
}

// captureProgressor records calls made by event adapters.
type captureProgressor struct {
	// calls stores normalized progress requests.
	calls []progressCall
	// hydrated stores hydrated player identifiers.
	hydrated []int64
	// flushed stores flushed player identifiers.
	flushed []int64
	// forgotten stores released player identifiers.
	forgotten []int64
}

// Progress records one ordinary trigger.
func (progressor *captureProgressor) Progress(_ context.Context, playerID int64, key string, amount int64) error {
	progressor.calls = append(progressor.calls, progressCall{playerID: playerID, key: key, amount: amount})
	return nil
}

// ProgressData records one metadata-bearing trigger.
func (progressor *captureProgressor) ProgressData(_ context.Context, playerID int64, key string, data string, amount int64) error {
	progressor.calls = append(progressor.calls, progressCall{playerID: playerID, key: key, data: data, amount: amount})
	return nil
}

// ProgressDaily records one daily trigger.
func (progressor *captureProgressor) ProgressDaily(_ context.Context, playerID int64, key string, amount int64) error {
	progressor.calls = append(progressor.calls, progressCall{playerID: playerID, key: key, amount: amount, daily: true})
	return nil
}

// HydratePlayer records one login hydration.
func (progressor *captureProgressor) HydratePlayer(_ context.Context, playerID int64) error {
	progressor.hydrated = append(progressor.hydrated, playerID)
	return nil
}

// SetTriggerProgress records one absolute trigger value.
func (progressor *captureProgressor) SetTriggerProgress(_ context.Context, playerID int64, key string, amount int64) error {
	progressor.calls = append(progressor.calls, progressCall{playerID: playerID, key: key, amount: amount})
	return nil
}

// FlushPlayer records one disconnect flush.
func (progressor *captureProgressor) FlushPlayer(_ context.Context, playerID int64) error {
	progressor.flushed = append(progressor.flushed, playerID)
	return nil
}

// ForgetPlayer records one disconnected cache release.
func (progressor *captureProgressor) ForgetPlayer(playerID int64) {
	progressor.forgotten = append(progressor.forgotten, playerID)
}

// assertCalls verifies one exact normalized trigger sequence.
func assertCalls(t *testing.T, actual []progressCall, expected []progressCall) {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("calls %#v, expected %#v", actual, expected)
	}
}

// TestRegistrationsAreCompleteAndUnique verifies every adapter has one bus topic.
func TestRegistrationsAreCompleteAndUnique(t *testing.T) {
	values := (&Subscriber{}).registrations()
	if len(values) != 32 {
		t.Fatalf("registrations %d", len(values))
	}
	seen := make(map[bus.Name]bool, len(values))
	for _, value := range values {
		if value.name == "" || value.handler == nil || seen[value.name] {
			t.Fatalf("invalid registration %#v", value)
		}
		seen[value.name] = true
	}
}

// TestFireworkChargedAdapter verifies one completed charge advances progression once.
func TestFireworkChargedAdapter(t *testing.T) {
	progressor := &captureProgressor{}
	subscriber := &Subscriber{engine: progressor}
	if err := subscriber.fireworkCharged(context.Background(), bus.Event{Payload: fireworkcharged.Payload{PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	assertCalls(t, progressor.calls, []progressCall{{playerID: 7, key: "furniture.firework.charged", amount: 1}})
}

// TestSessionAndModerationAdapters verifies lifecycle and self-moderation mappings.
func TestSessionAndModerationAdapters(t *testing.T) {
	progressor := &captureProgressor{}
	subscriber := &Subscriber{engine: progressor}
	ctx := context.Background()
	if err := subscriber.authenticated(ctx, bus.Event{Payload: playerauthenticated.Payload{PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	if err := subscriber.selfModSettings(ctx, bus.Event{Payload: roomsettings.Payload{ActorID: 7}}); err != nil {
		t.Fatal(err)
	}
	if err := subscriber.selfModWordFilter(ctx, bus.Event{Payload: roomwordfilter.Payload{ActorID: 7}}); err != nil {
		t.Fatal(err)
	}
	if err := subscriber.selfModMuted(ctx, bus.Event{Payload: roommuted.Payload{ActorID: 7}}); err != nil {
		t.Fatal(err)
	}
	if err := subscriber.selfModKicked(ctx, bus.Event{Payload: roomkicked.Payload{ActorID: 7}}); err != nil {
		t.Fatal(err)
	}
	if err := subscriber.subscriptionPayday(ctx, bus.Event{Payload: subscriptionpayday.Payload{PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	if err := subscriber.staffPicked(ctx, bus.Event{Payload: roomstaffpicked.Payload{OwnerPlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	if err := subscriber.disconnected(ctx, bus.Event{Payload: playerdisconnected.Payload{PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	expected := []progressCall{{7, "player.login.streak", "", 1, true}, {7, "selfmod.door_mode", "", 1, false}, {7, "selfmod.walkthrough", "", 1, false}, {7, "selfmod.chat_scroll_speed", "", 1, false}, {7, "selfmod.chat_hear_range", "", 1, false}, {7, "selfmod.chat_flood_filter", "", 1, false}, {7, "selfmod.room_filter", "", 1, false}, {7, "selfmod.mute", "", 1, false}, {7, "selfmod.kick", "", 1, false}, {7, "subscription.hc.month", "", 1, false}, {7, "staffpick.received", "", 1, false}}
	assertCalls(t, progressor.calls, expected)
	if !reflect.DeepEqual(progressor.hydrated, []int64{7}) || !reflect.DeepEqual(progressor.flushed, []int64{7}) || !reflect.DeepEqual(progressor.forgotten, []int64{7}) {
		t.Fatalf("lifecycle hydrated=%v flushed=%v forgotten=%v", progressor.hydrated, progressor.flushed, progressor.forgotten)
	}
}
