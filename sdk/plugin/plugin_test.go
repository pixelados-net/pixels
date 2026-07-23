package plugin

import "testing"

// TestMetadataValid verifies complete plugin metadata is valid.
func TestMetadataValid(t *testing.T) {
	metadata := Metadata{
		Name:       "example",
		Version:    "v0.1.0",
		Author:     "Pixels QA",
		SDKVersion: SDKVersion,
	}

	if !metadata.Valid() {
		t.Fatal("expected complete metadata to be valid")
	}
}

// TestMetadataValidRequiresFields verifies plugin metadata requires identity fields.
func TestMetadataValidRequiresFields(t *testing.T) {
	metadata := Metadata{Name: "example"}

	if metadata.Valid() {
		t.Fatal("expected incomplete metadata to be invalid")
	}
}
