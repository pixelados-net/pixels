package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewAppBuilds verifies the dependency graph can be constructed.
func TestNewAppBuilds(t *testing.T) {
	setI18NPathForTest(t)
	t.Setenv("PIXELS_CURRENCY_TYPES", "-1:credits")
	setFigureDataPathForTest(t)
	app := newApp()

	if app == nil {
		t.Fatal("expected app")
	}
	if err := app.Err(); err != nil {
		t.Fatalf("build dependency graph: %v", err)
	}
}

// setFigureDataPathForTest points app construction at a minimal figure catalog.
func setFigureDataPathForTest(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "figuredata.xml")
	data := []byte(`<figuredata><colors><palette id="1"><color id="1" club="0" selectable="1"/></palette></colors><sets><settype type="hd" paletteid="1"><set id="180" gender="U" club="0" selectable="1"/></settype></sets></figuredata>`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write figure data: %v", err)
	}
	t.Setenv("PIXELS_FIGURE_DATA_PATH", path)
}

// TestOptionsBuilds verifies dependency graph options are registered.
func TestOptionsBuilds(t *testing.T) {
	options := options()

	if len(options) != 37 {
		t.Fatalf("expected thirty-seven options, got %d", len(options))
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
