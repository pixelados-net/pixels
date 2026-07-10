package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewAppBuilds verifies the dependency graph can be constructed.
func TestNewAppBuilds(t *testing.T) {
	setI18NPathForTest(t)
	setCurrencyPathForTest(t)
	app := newApp()

	if app == nil {
		t.Fatal("expected app")
	}
	if err := app.Err(); err != nil {
		t.Fatalf("build dependency graph: %v", err)
	}
}

// setCurrencyPathForTest points app construction at a minimal currency catalog.
func setCurrencyPathForTest(t *testing.T) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "currencies.json")
	if err := os.WriteFile(path, []byte(`[{"type":-1,"key":"credits"}]`), 0o600); err != nil {
		t.Fatalf("write currency catalog: %v", err)
	}
	t.Setenv("PIXELS_CURRENCY_CATALOG_PATH", path)
}

// TestOptionsBuilds verifies dependency graph options are registered.
func TestOptionsBuilds(t *testing.T) {
	options := options()

	if len(options) != 21 {
		t.Fatalf("expected twenty-one options, got %d", len(options))
	}
}

// setI18NPathForTest points app construction at an empty test catalog.
func setI18NPathForTest(t *testing.T) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "translations.json")
	if err := os.WriteFile(path, []byte(`{"locales":{}}`), 0o600); err != nil {
		t.Fatalf("write i18n catalog: %v", err)
	}
	t.Setenv("PIXELS_I18N_PATH", path)
}
