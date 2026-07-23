package record

import (
	"context"
	"testing"
)

// TestEnumsRejectWireOnlyAndUnknownValues verifies durable enum boundaries.
func TestEnumsRejectWireOnlyAndUnknownValues(t *testing.T) {
	if !Regular.Valid() || !Private.Valid() || State(3).Valid() {
		t.Fatal("unexpected group state validation")
	}
	if !Owner.Valid() || !Member.Valid() || Requested.Valid() {
		t.Fatal("unexpected role validation")
	}
	if !Everyone.Valid() || !Owners.Valid() || Policy(4).Valid() {
		t.Fatal("unexpected policy validation")
	}
	if !ThreadOpen.Valid() || !ThreadHiddenAdmin.Valid() || ThreadState(2).Valid() {
		t.Fatal("unexpected thread state validation")
	}
}

// TestAuditAttributionRoundTrip verifies request-scoped mutation attribution.
func TestAuditAttributionRoundTrip(t *testing.T) {
	ctx := WithAudit(context.Background(), 7, "qa mutation")
	audit, found := AuditFromContext(ctx)
	if !found || audit.ActorPlayerID != 7 || audit.Reason != "qa mutation" {
		t.Fatalf("audit=%#v found=%v", audit, found)
	}
	if _, found = AuditFromContext(context.Background()); found {
		t.Fatal("unexpected attribution in empty context")
	}
}

// TestFurnitureReturnCountReportsExactBatch verifies cleanup confirmation uses the projected batch.
func TestFurnitureReturnCountReportsExactBatch(t *testing.T) {
	result := FurnitureReturn{Items: []ReturnedFurniture{{ItemID: 39}, {ItemID: 40}}}
	if result.Count() != 2 {
		t.Fatalf("expected two returned items, got %d", result.Count())
	}
	if (FurnitureReturn{}).Count() != 0 {
		t.Fatal("empty cleanup reported returned furniture")
	}
}
