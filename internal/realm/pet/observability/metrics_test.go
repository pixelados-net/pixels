package observability

import (
	"testing"
	"time"
)

// TestMetricsSnapshotReportsBoundedCounters verifies every public metric family.
func TestMetricsSnapshotReportsBoundedCounters(t *testing.T) {
	metrics := New()
	metrics.RecordOperation(OperationPlace, ResultSuccess)
	metrics.RecordDecision(DecisionNeed)
	metrics.RecordAction(46, ResultRejected)
	metrics.RecordPath(PathBlocked)
	metrics.RecordStatFlush(ResultFailed)
	metrics.RecordBreeding(BreedingConfirm, ResultSuccess)
	metrics.RecordProduct(ProductFood, ResultSuccess)
	metrics.AddRoom(1)
	metrics.ObserveRoomLoad(75 * time.Microsecond)
	metrics.ObserveInventoryList(2 * time.Millisecond)
	metrics.ObserveBehaviorDue(50 * time.Millisecond)
	metrics.ObserveTransaction(time.Millisecond)
	snapshot := metrics.Snapshot()
	if snapshot.Operations[OperationPlace].Success != 1 || snapshot.BehaviorDecisions[DecisionNeed] != 1 || snapshot.Actions[46].Rejected != 1 {
		t.Fatalf("operation snapshot=%+v", snapshot)
	}
	if snapshot.PathResults[PathBlocked] != 1 || snapshot.StatFlush.Failed != 1 || snapshot.Breeding[BreedingConfirm].Success != 1 || snapshot.ProductUses[ProductFood].Success != 1 {
		t.Fatalf("domain snapshot=%+v", snapshot)
	}
	if snapshot.RoomCount != 1 || snapshot.RoomLoad.Buckets[1] != 1 || snapshot.InventoryList.Count != 1 || snapshot.BehaviorDue.Buckets[7] != 1 || snapshot.Transaction.Count != 1 {
		t.Fatalf("timing snapshot=%+v", snapshot)
	}
}

// TestMetricsRejectsOutOfRangeIndexes verifies malformed labels cannot panic.
func TestMetricsRejectsOutOfRangeIndexes(t *testing.T) {
	metrics := New()
	metrics.RecordOperation(operationCount, resultCount)
	metrics.RecordDecision(decisionCount)
	metrics.RecordAction(-1, ResultSuccess)
	metrics.RecordAction(47, ResultSuccess)
	metrics.RecordPath(pathResultCount)
	metrics.RecordBreeding(breedingOperationCount, ResultSuccess)
	metrics.RecordProduct(productKindCount, ResultSuccess)
	if snapshot := metrics.Snapshot(); snapshot != (Snapshot{}) {
		t.Fatalf("unexpected counters %+v", snapshot)
	}
}
