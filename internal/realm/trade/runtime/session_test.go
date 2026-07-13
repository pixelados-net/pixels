package runtime

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestSessionMutationsResetAgreement verifies offer changes invalidate both agreement phases.
func TestSessionMutationsResetAgreement(t *testing.T) {
	session := &Session{First: Participant{PlayerID: 1}, Second: Participant{PlayerID: 2}}
	session.SetAccepted(1, true)
	session.SetAccepted(2, true)
	if both, ok := session.Confirm(1); both || !ok {
		t.Fatal("expected first confirmation")
	}
	if !session.AddItem(1, 10) {
		t.Fatal("add")
	}
	first, second := session.Snapshot()
	if first.Accepted || second.Accepted || first.Confirmed || second.Confirmed {
		t.Fatal("offer mutation did not reset agreement")
	}
}

// TestConcurrentConfirmationClaimsOneSettlement verifies exactly one confirmer owns persistence.
func TestConcurrentConfirmationClaimsOneSettlement(t *testing.T) {
	session := &Session{First: Participant{PlayerID: 1}, Second: Participant{PlayerID: 2}}
	session.SetAccepted(1, true)
	session.SetAccepted(2, true)
	start := make(chan struct{})
	var ready atomic.Int32
	var wait sync.WaitGroup
	for _, playerID := range []int64{1, 2} {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			if claimed, updated := session.Confirm(playerID); claimed && updated {
				ready.Add(1)
			}
		}()
	}
	close(start)
	wait.Wait()
	if ready.Load() != 1 {
		t.Fatalf("settlement claims = %d", ready.Load())
	}
	if session.TryCancel() {
		t.Fatal("cancellation claimed a settling session")
	}
}
