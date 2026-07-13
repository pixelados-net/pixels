package runtime

import (
	"sync"
	"testing"
)

// TestRegistryPreventsDuplicateSessionsAndItems verifies atomic participant and item indexes.
func TestRegistryPreventsDuplicateSessionsAndItems(t *testing.T) {
	registry := NewRegistry()
	session := &Session{RoomID: 9, First: Participant{PlayerID: 1}, Second: Participant{PlayerID: 2}}
	if !registry.Start(session) || registry.Start(&Session{First: Participant{PlayerID: 1}, Second: Participant{PlayerID: 3}}) {
		t.Fatal("unexpected session registration")
	}
	if !registry.Stage(1, 10) || registry.Stage(2, 10) || !registry.Contains(10) {
		t.Fatal("unexpected staging result")
	}
	if !registry.Unstage(1, 10) || registry.Contains(10) {
		t.Fatal("expected item release")
	}
	if closed, found := registry.Close(2); !found || closed != session || registry.ActiveCount() != 0 {
		t.Fatal("expected complete session cleanup")
	}
}

// TestRegistryConcurrentStageHasSingleWinner verifies staged item ownership under contention.
func TestRegistryConcurrentStageHasSingleWinner(t *testing.T) {
	registry := NewRegistry()
	if !registry.Start(&Session{First: Participant{PlayerID: 1}, Second: Participant{PlayerID: 2}}) {
		t.Fatal("start")
	}
	var wait sync.WaitGroup
	wins := make(chan bool, 2)
	for _, playerID := range []int64{1, 2} {
		wait.Add(1)
		go func(id int64) { defer wait.Done(); wins <- registry.Stage(id, 99) }(playerID)
	}
	wait.Wait()
	close(wins)
	count := 0
	for won := range wins {
		if won {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected one winner, got %d", count)
	}
}

// TestStageManyDoesNotPartiallyMutate verifies a conflicting batch stages no items.
func TestStageManyDoesNotPartiallyMutate(t *testing.T) {
	registry := NewRegistry()
	if !registry.Start(&Session{First: Participant{PlayerID: 1}, Second: Participant{PlayerID: 2}}) || !registry.Stage(2, 11) {
		t.Fatal("fixture setup")
	}
	if registry.StageMany(1, []int64{10, 11}) || registry.Contains(10) {
		t.Fatal("conflicting batch partially staged")
	}
	participant, _ := registry.byPlayer[1].Participant(1)
	if len(participant.Items) != 0 {
		t.Fatalf("unexpected offer %#v", participant.Items)
	}
}

// BenchmarkRegistryContains measures the staged-item hot path.
func BenchmarkRegistryContains(b *testing.B) {
	registry := NewRegistry()
	registry.staged[42] = 1
	b.ReportAllocs()
	for b.Loop() {
		if !registry.Contains(42) {
			b.Fatal("missing staged item")
		}
	}
}
