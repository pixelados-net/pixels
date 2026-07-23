// Package tests verifies all trigger match categories.
package tests

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// TestMatcher verifies event, actor, text, and selected-item matching.
func TestMatcher(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		parameters configuration.Parameters
		targets    []record.Target
		event      trigger.Event
		want       bool
	}{
		{name: "say substring", key: "wf_trg_says_something", parameters: configuration.Parameters{Text: "hola"}, event: trigger.Event{Kind: trigger.Say, RoomID: 1, ActorKind: trigger.ActorPlayer, Message: "HOLA mundo"}, want: true},
		{name: "walk target", key: "wf_trg_walks_on_furni", targets: []record.Target{{ItemID: 8}}, event: trigger.Event{Kind: trigger.WalkOn, RoomID: 1, ActorKind: trigger.ActorPet, SourceItem: 8}, want: true},
		{name: "wrong actor", key: "wf_trg_enter_room", event: trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorBot}, want: false},
		{name: "system game", key: "wf_trg_game_starts", event: trigger.Event{Kind: trigger.GameStarted, RoomID: 1}, want: true},
	}
	manifest := map[string]registry.Descriptor{}
	for _, descriptor := range registry.CanonicalManifest() {
		manifest[descriptor.Key] = descriptor
	}
	matcher := trigger.New()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mode := int32(0)
			if len(test.targets) > 0 {
				mode = 1
			}
			node := &configuration.Node{RoomID: 1, Descriptor: manifest[test.key], Parameters: test.parameters, Targets: test.targets, SelectionMode: mode}
			if got := matcher.Match(node, test.event); got != test.want {
				t.Fatalf("Match() = %t, want %t", got, test.want)
			}
		})
	}
}

// TestAllCanonicalTriggersMatch verifies every descriptor maps to a concrete event kind.
func TestAllCanonicalTriggersMatch(t *testing.T) {
	kinds := map[string]trigger.Kind{
		"wf_trg_enter_room": trigger.EnterRoom, "wf_trg_says_something": trigger.Say,
		"wf_trg_walks_on_furni": trigger.WalkOn, "wf_trg_walks_off_furni": trigger.WalkOff,
		"wf_trg_state_changed": trigger.StateChanged, "wf_trg_collision": trigger.Collision,
		"wf_trg_periodically": trigger.Periodic, "wf_trg_period_long": trigger.PeriodicLong,
		"wf_trg_at_given_time": trigger.AtTime, "wf_trg_at_time_long": trigger.AtTimeLong,
		"wf_trg_game_starts": trigger.GameStarted, "wf_trg_game_ends": trigger.GameEnded,
		"wf_trg_score_achieved": trigger.ScoreAchieved, "wf_trg_bot_reached_stf": trigger.BotReachedFurniture,
		"wf_trg_bot_reached_avtr": trigger.BotReachedAvatar, "wf_trg_game_team_win": trigger.TeamWon,
		"wf_trg_game_team_lose": trigger.TeamLost,
	}
	matcher := trigger.New()
	count := 0
	for _, descriptor := range registry.CanonicalManifest() {
		if descriptor.Family != registry.FamilyTrigger {
			continue
		}
		count++
		kind, found := kinds[descriptor.Key]
		if !found {
			t.Fatalf("missing trigger event mapping for %s", descriptor.Key)
		}
		node := &configuration.Node{RoomID: 1, Descriptor: descriptor, Parameters: configuration.Parameters{Text: "hello", Name: "bot", Values: []int32{5}}, SelectionMode: 1, Targets: []record.Target{{ItemID: 8, SpriteID: 9}}}
		if descriptor.Key == "wf_trg_enter_room" {
			node.Parameters.Text = ""
		}
		event := trigger.Event{Kind: kind, RoomID: 1, ActorKind: trigger.ActorPlayer, ActorID: 4, PlayerID: 4, Username: "bot", Message: "hello world", SourceItem: 8, SourceSprite: 9, PreviousScore: 4, Score: 5}
		if descriptor.Actor == registry.ActorBot {
			event.ActorKind = trigger.ActorBot
		}
		if !matcher.Match(node, event) {
			t.Fatalf("trigger %s did not match event %d", descriptor.Key, kind)
		}
	}
	if count != 17 {
		t.Fatalf("trigger count=%d", count)
	}
}

// BenchmarkMatcher verifies candidate matching remains allocation free.
func BenchmarkMatcher(b *testing.B) {
	descriptor := registry.Descriptor{Key: "wf_trg_walks_on_furni", Family: registry.FamilyTrigger, Actor: registry.ActorUnit}
	node := &configuration.Node{RoomID: 1, Descriptor: descriptor, Targets: []record.Target{{ItemID: 8}}, SelectionMode: 1}
	event := trigger.Event{Kind: trigger.WalkOn, RoomID: 1, ActorKind: trigger.ActorPlayer, SourceItem: 8}
	matcher := trigger.New()
	b.ReportAllocs()
	for range b.N {
		if !matcher.Match(node, event) {
			b.Fatal("did not match")
		}
	}
}
