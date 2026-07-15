package database

import (
	"testing"
	"time"
)

// TestMergeExpiryPreservesLongestOrPermanentWindow verifies overlap aggregation.
func TestMergeExpiryPreservesLongestOrPermanentWindow(t *testing.T) {
	now := time.Now()
	short := now.Add(time.Hour)
	long := now.Add(2 * time.Hour)
	permanent, expiry := mergeExpiry(false, &short, &long)
	if permanent || expiry == nil || *expiry != long {
		t.Fatalf("permanent=%v expiry=%v", permanent, expiry)
	}
	permanent, expiry = mergeExpiry(false, &long, nil)
	if !permanent || expiry != nil {
		t.Fatalf("permanent=%v expiry=%v", permanent, expiry)
	}
}
