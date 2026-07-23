package i18n

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

// TestLoadCatalogReadsJSON verifies catalog loading from disk.
func TestLoadCatalogReadsJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "translations.json")
	writeFile(t, path, `{"version":1,"locales":{"es":{"hello":"Hola"}}}`)

	catalog, err := LoadCatalog(Config{Path: path, DefaultLocale: "es"}, zap.NewNop())
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	if got := catalog.Default("hello"); got != "Hola" {
		t.Fatalf("expected Hola, got %q", got)
	}
}

// TestLoadCatalogAllowsMissingFile verifies missing files fall back to raw keys.
func TestLoadCatalogAllowsMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")

	catalog, err := LoadCatalog(Config{Path: path, DefaultLocale: "es"}, zap.NewNop())
	if err != nil {
		t.Fatalf("load missing catalog: %v", err)
	}

	if got := catalog.Default("hello"); got != "hello" {
		t.Fatalf("expected key fallback, got %q", got)
	}
}

// TestLoadCatalogRejectsInvalidJSON verifies corrupt files fail startup.
func TestLoadCatalogRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "translations.json")
	writeFile(t, path, `{`)

	if _, err := LoadCatalog(Config{Path: path}, zap.NewNop()); err == nil {
		t.Fatal("expected parse error")
	}
}

// writeFile writes a test file.
func writeFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
