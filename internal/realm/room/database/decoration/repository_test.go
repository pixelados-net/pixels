package decoration

import (
	"strings"
	"testing"
)

// TestPlacePostItSQLQualifiesMutationVersion verifies joined tables cannot make optimistic state ambiguous.
func TestPlacePostItSQLQualifiesMutationVersion(t *testing.T) {
	if !strings.Contains(placePostItSQL, "version=fi.version+1") {
		t.Fatalf("post-it query must qualify the furniture item version: %s", placePostItSQL)
	}
	if strings.Contains(placePostItSQL, "version=version+1") {
		t.Fatalf("post-it query contains an ambiguous version expression: %s", placePostItSQL)
	}
	if !strings.Contains(placePostItSQL, "extra_data=$5") {
		t.Fatalf("post-it query must persist renderable initial data: %s", placePostItSQL)
	}
}
