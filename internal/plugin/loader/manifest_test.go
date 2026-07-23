package loader

import (
	"path/filepath"
	"testing"

	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
)

// TestValidMetadataRejectsUnsafeManifests verifies every embedded identity rule.
func TestValidMetadataRejectsUnsafeManifests(t *testing.T) {
	valid := fixtureMetadata("valid-plugin")
	tests := []sdkplugin.Metadata{
		{},
		{Name: "UPPER", Version: "1.0.0", Author: "QA", SDKVersion: sdkplugin.SDKVersion},
		{Name: "-leading", Version: "1.0.0", Author: "QA", SDKVersion: sdkplugin.SDKVersion},
		{Name: "trailing-", Version: "1.0.0", Author: "QA", SDKVersion: sdkplugin.SDKVersion},
		{Name: "bad version", Version: "nope", Author: "QA", SDKVersion: sdkplugin.SDKVersion},
		{Name: "self", Version: "1.0.0", Author: "QA", SDKVersion: sdkplugin.SDKVersion, Dependencies: []string{"self"}},
		{Name: "duplicates", Version: "1.0.0", Author: "QA", SDKVersion: sdkplugin.SDKVersion, Dependencies: []string{"base", "base"}},
	}
	if !validMetadata(valid) {
		t.Fatal("expected valid metadata")
	}
	for _, metadata := range tests {
		if validMetadata(metadata) {
			t.Fatalf("expected invalid metadata: %+v", metadata)
		}
	}
}

// TestDiscoverAllowsMissingDirectory verifies plugins remain optional.
func TestDiscoverAllowsMissingDirectory(t *testing.T) {
	paths, err := discover(filepath.Join(t.TempDir(), "missing"))
	if err != nil || len(paths) != 0 {
		t.Fatalf("expected empty discovery, paths=%v err=%v", paths, err)
	}
}

// TestNativeOpenerRejectsMissingObject verifies native errors remain isolated values.
func TestNativeOpenerRejectsMissingObject(t *testing.T) {
	opener := NewNative()
	if _, err := opener.Open(filepath.Join(t.TempDir(), "missing.so")); err == nil {
		t.Fatal("expected missing native object error")
	}
}
