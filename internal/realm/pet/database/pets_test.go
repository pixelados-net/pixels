package database

import "testing"

// TestOperationKindClassifiesIdempotentGrants verifies durable support labels.
func TestOperationKindClassifiesIdempotentGrants(t *testing.T) {
	tests := map[string]string{
		"breeding:9:1":  "breeding",
		"package:42":    "package",
		"seed:42":       "grant",
		"admin:request": "grant",
	}
	for key, expected := range tests {
		if actual := operationKind(key); actual != expected {
			t.Fatalf("operation %q: expected %q, got %q", key, expected, actual)
		}
	}
}
