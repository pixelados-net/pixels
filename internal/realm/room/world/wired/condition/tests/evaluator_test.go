// Package tests verifies canonical WIRED condition semantics.
package tests

import (
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/condition"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// view implements focused condition facts.
type view struct {
	// users stores player occupancy.
	users int
	// predicate stores the shared behavior result.
	predicate bool
	// valid stores whether actor domain exists.
	valid bool
	// err stores an injected dependency error.
	err error
}

// UserCount returns configured player occupancy.
func (value view) UserCount() int { return value.users }

// UnitOn returns the shared predicate.
func (value view) UnitOn(int64) (bool, error) { return value.predicate, value.err }

// ActorOn returns the shared predicate and domain.
func (value view) ActorOn(trigger.Event, int64) (bool, bool, error) {
	return value.predicate, value.valid, value.err
}

// Stacked returns the shared predicate.
func (value view) Stacked(int64) (bool, error) { return value.predicate, value.err }

// SnapshotMatches returns the shared predicate.
func (value view) SnapshotMatches(int64, record.Snapshot, []int32) (bool, error) {
	return value.predicate, value.err
}

// ActorTeam returns the shared predicate and domain.
func (value view) ActorTeam(int64, int32) (bool, bool, error) {
	return value.predicate, value.valid, value.err
}

// ActorGroup returns the shared predicate and domain.
func (value view) ActorGroup(int64) (bool, bool, error) {
	return value.predicate, value.valid, value.err
}

// WearingBadge returns the shared predicate and domain.
func (value view) WearingBadge(int64, string) (bool, bool, error) {
	return value.predicate, value.valid, value.err
}

// WearingEffect returns the shared predicate and domain.
func (value view) WearingEffect(int64, int32) (bool, bool, error) {
	return value.predicate, value.valid, value.err
}

// HasHanditem returns the shared predicate and domain.
func (value view) HasHanditem(int64, int32) (bool, bool, error) {
	return value.predicate, value.valid, value.err
}

// ValidMoves returns the shared compatibility simulation result.
func (value view) ValidMoves([]*configuration.Node, trigger.Event) (bool, error) {
	return value.predicate, value.err
}

// TestNegativeFailsClosed verifies invalid/error domains never pass negation.
func TestNegativeFailsClosed(t *testing.T) {
	evaluator := condition.New()
	node := nodeFor("wf_cnd_not_in_group")
	result, err := evaluator.Evaluate(node, condition.Context{Event: trigger.Event{PlayerID: 8}}, view{valid: false})
	if err != nil || result.Pass {
		t.Fatalf("invalid negative result = %+v, %v", result, err)
	}
	injected := errors.New("dependency")
	result, err = evaluator.Evaluate(node, condition.Context{Event: trigger.Event{PlayerID: 8}}, view{valid: true, err: injected})
	if !errors.Is(err, injected) || result.Pass {
		t.Fatalf("error negative result = %+v, %v", result, err)
	}
}

// TestRangesAndTime verifies inclusive occupancy and monotonic elapsed checks.
func TestRangesAndTime(t *testing.T) {
	evaluator := condition.New()
	rangeNode := nodeFor("wf_cnd_user_count_in")
	rangeNode.Parameters.Values = []int32{2, 4}
	result, err := evaluator.Evaluate(rangeNode, condition.Context{}, view{users: 4})
	if err != nil || !result.Pass {
		t.Fatalf("range result = %+v, %v", result, err)
	}
	timeNode := nodeFor("wf_cnd_time_more_than")
	timeNode.Parameters.Duration = time.Second
	now := time.Now()
	result, err = evaluator.Evaluate(timeNode, condition.Context{Now: now, ResetAt: now.Add(-2 * time.Second)}, view{})
	if err != nil || !result.Pass {
		t.Fatalf("time result = %+v, %v", result, err)
	}
}

// TestAllCanonicalConditionsEvaluateValid verifies every descriptor reaches a concrete predicate.
func TestAllCanonicalConditionsEvaluateValid(t *testing.T) {
	evaluator := condition.New()
	now := time.Unix(1_700_000_000, 0)
	count := 0
	for _, descriptor := range registry.CanonicalManifest() {
		if descriptor.Family != registry.FamilyCondition {
			continue
		}
		count++
		node := &configuration.Node{
			Descriptor: descriptor, SelectionMode: 1, Parameters: configuration.Parameters{Values: []int32{1, 4, 1}, Text: "ACH_TEST", Duration: time.Second},
			Targets: []record.Target{{ItemID: 8, SpriteID: 9, Snapshot: record.Snapshot{Present: true}}},
		}
		if descriptor.Key == "wf_cnd_date_rng_active" {
			node.Parameters.Values = []int32{int32(now.Add(-time.Second).Unix()), int32(now.Add(time.Second).Unix())}
		}
		result, err := evaluator.Evaluate(node, condition.Context{Event: trigger.Event{ActorID: 4, PlayerID: 4, SourceItem: 8, SourceSprite: 9}, Now: now, ResetAt: now.Add(-2 * time.Second)}, view{users: 2, predicate: true, valid: true})
		if err != nil || !result.Valid {
			t.Fatalf("condition %s result=%+v err=%v", descriptor.Key, result, err)
		}
	}
	if count != 24 {
		t.Fatalf("condition count=%d", count)
	}
}

// nodeFor creates one test node from the canonical manifest.
func nodeFor(key string) *configuration.Node {
	manifest := append(registry.CanonicalManifest(), registry.CompatibilityManifest()...)
	for _, descriptor := range manifest {
		if descriptor.Key == key {
			return &configuration.Node{Descriptor: descriptor}
		}
	}
	panic(key)
}
