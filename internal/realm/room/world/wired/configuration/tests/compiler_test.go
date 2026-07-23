// Package tests verifies WIRED configuration compilation.
package tests

import (
	"errors"
	"testing"
	"time"

	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// TestCompileBuildsStableStack verifies families and stack add-ons.
func TestCompileBuildsStableStack(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	records := []record.Config{
		{ItemID: 3, RoomID: 1, Interaction: "wf_act_show_message", X: 2, Y: 3, StringParam: "hello"},
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 2, Y: 3},
		{ItemID: 2, RoomID: 1, Interaction: "wf_xtra_or_eval", X: 2, Y: 3},
	}
	generation, err := compiler.Compile(1, 8, records)
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	stack := generation.Stacks[configuration.Point{X: 2, Y: 3}]
	if stack == nil || !stack.Or || len(stack.Triggers) != 1 || len(stack.Effects) != 1 || stack.Effects[0].ItemID != 3 {
		t.Fatalf("unexpected stack: %+v", stack)
	}
}

// TestCompileRejectsEveryStrictBehaviorBoundary verifies editor-specific schema constraints.
func TestCompileRejectsEveryStrictBehaviorBoundary(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	cases := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_says_something", StringParam: " "},
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_periodically", IntParams: []int32{0}},
		{ItemID: 1, RoomID: 1, Interaction: "wf_cnd_user_count_in", IntParams: []int32{5, 4}},
		{ItemID: 1, RoomID: 1, Interaction: "wf_act_join_team", IntParams: []int32{5}},
		{ItemID: 1, RoomID: 1, Interaction: "wf_cnd_date_rng_active", IntParams: []int32{2, 1}},
		{ItemID: 1, RoomID: 1, Interaction: "wf_act_mute_triggerer", IntParams: []int32{1441}},
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", SelectionMode: 1},
		{ItemID: 1, RoomID: 1, Interaction: "wf_act_toggle_state", SelectionMode: 1, Targets: []record.Target{{ItemID: -1}}},
		{ItemID: 1, RoomID: 1, Interaction: "wf_act_toggle_state", SelectionMode: 1, Targets: []record.Target{{ItemID: 2}, {ItemID: 2}}},
	}
	for _, stored := range cases {
		if _, compileErr := compiler.CompileNode(stored); !errors.Is(compileErr, configuration.ErrInvalid) {
			t.Errorf("interaction=%s error=%v", stored.Interaction, compileErr)
		}
	}
}

// TestCompileParsesCompatibilityParametersOutsideHotPath verifies normalized bot, badge, numeric, and timer data.
func TestCompileParsesCompatibilityParametersOutsideHotPath(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	badge, err := compiler.CompileNode(record.Config{ItemID: 1, RoomID: 1, Interaction: "wf_cnd_wearing_badge", StringParam: " badge_qa "})
	if err != nil || badge.Parameters.Text != "BADGE_QA" {
		t.Fatalf("badge=%+v err=%v", badge, err)
	}
	bot, err := compiler.CompileNode(record.Config{ItemID: 2, RoomID: 1, Interaction: "wf_act_bot_talk", StringParam: " Helper\tHello there "})
	if err != nil || bot.Parameters.Name != "Helper" || bot.Parameters.Message != "Hello there" {
		t.Fatalf("bot=%+v err=%v", bot, err)
	}
	numeric, err := compiler.CompileNode(record.Config{ItemID: 3, RoomID: 1, Interaction: "wf_act_give_effect", StringParam: "42"})
	if err != nil || numeric.Parameters.Number != 42 {
		t.Fatalf("numeric=%+v err=%v", numeric, err)
	}
	long, err := compiler.CompileNode(record.Config{ItemID: 4, RoomID: 1, Interaction: "wf_trg_period_long", IntParams: []int32{2}})
	if err != nil || long.Parameters.Duration != 10*time.Second {
		t.Fatalf("long timer=%+v err=%v", long, err)
	}
	if _, err = compiler.CompileNode(record.Config{ItemID: 5, RoomID: 1, Interaction: "wf_act_give_handitem", StringParam: "invalid"}); !errors.Is(err, configuration.ErrInvalid) {
		t.Fatalf("invalid numeric error=%v", err)
	}
}

// TestCompileRejectsInvalidSelection verifies required target validation.
func TestCompileRejectsInvalidSelection(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	_, err = compiler.CompileNode(record.Config{ItemID: 1, RoomID: 1, Interaction: "wf_act_toggle_state"})
	if !errors.Is(err, configuration.ErrInvalid) {
		t.Fatalf("CompileNode() error = %v", err)
	}
}

// TestCompileGivesUnseenSelectorPrecedence verifies ambiguous imported stacks remain deterministic.
func TestCompileGivesUnseenSelectorPrecedence(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	generation, err := compiler.Compile(1, 1, []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_xtra_random", X: 2, Y: 2},
		{ItemID: 2, RoomID: 1, Interaction: "wf_xtra_unseen", X: 2, Y: 2},
	})
	if err != nil {
		t.Fatalf("Compile() error = %v", err)
	}
	stack := generation.Stacks[configuration.Point{X: 2, Y: 2}]
	if stack == nil || !stack.Unseen || stack.Random {
		t.Fatalf("unexpected selectors %+v", stack)
	}
}

// BenchmarkCompileNode measures configuration compilation outside the hot path.
func BenchmarkCompileNode(b *testing.B) {
	registered, err := registry.Canonical()
	if err != nil {
		b.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	stored := record.Config{ItemID: 1, RoomID: 1, Interaction: "wf_trg_says_something", StringParam: "hello"}
	b.ReportAllocs()
	for range b.N {
		if _, compileErr := compiler.CompileNode(stored); compileErr != nil {
			b.Fatal(compileErr)
		}
	}
}
