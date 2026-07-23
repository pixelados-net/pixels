package entry

import (
	"testing"
	"time"
)

// TestTrustStoreConsumesEntriesOnce verifies trusted entry lifetime and one-time use.
func TestTrustStoreConsumesEntriesOnce(t *testing.T) {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	var store TrustStore
	if !store.Grant(7, 9, now.Add(time.Second)) {
		t.Fatal("expected trusted entry grant")
	}
	if !store.Consume(7, 9, now) {
		t.Fatal("expected trusted entry consumption")
	}
	if store.Consume(7, 9, now) || store.Len() != 0 {
		t.Fatal("expected one-time trusted entry removal")
	}
	store.Grant(7, 9, now.Add(-time.Second))
	if store.Consume(7, 9, now) {
		t.Fatal("expected expired trusted entry rejection")
	}
}
