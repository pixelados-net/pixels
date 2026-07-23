package build

import "testing"

// TestDefaultInfo verifies the local development build metadata.
func TestDefaultInfo(t *testing.T) {
	info := DefaultInfo()

	if info.Name != Name {
		t.Fatalf("expected name %q, got %q", Name, info.Name)
	}

	if info.Version != Version {
		t.Fatalf("expected default version, got %q", info.Version)
	}
}

// TestNewInfoKeepsSemanticVersionAndCommit verifies independent build metadata.
func TestNewInfoKeepsSemanticVersionAndCommit(t *testing.T) {
	info := NewInfo("pixels", "v1.2.3", "1234567890abcdef")

	if info.Version != "v1.2.3" {
		t.Fatalf("expected semantic version, got %q", info.Version)
	}

	if info.Commit != "12345678" {
		t.Fatalf("expected short commit, got %q", info.Commit)
	}
}

// TestShortCommitKeepsShortValues verifies short commit values are stable.
func TestShortCommitKeepsShortValues(t *testing.T) {
	commit := ShortCommit("dev")

	if commit != "dev" {
		t.Fatalf("expected dev commit, got %q", commit)
	}
}
