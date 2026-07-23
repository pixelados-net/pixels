package i18n

import "testing"

// TestCatalogTranslatesDefault verifies default locale lookup.
func TestCatalogTranslatesDefault(t *testing.T) {
	catalog := NewCatalog(Config{DefaultLocale: "es", FallbackLocale: "en"}, testEntries())

	if got := catalog.Default("hello"); got != "Hola" {
		t.Fatalf("expected Hola, got %q", got)
	}
}

// TestCatalogFallsBack verifies missing locale fallback.
func TestCatalogFallsBack(t *testing.T) {
	catalog := NewCatalog(Config{DefaultLocale: "es", FallbackLocale: "en"}, testEntries())

	if got := catalog.T("fr", "hello"); got != "Hello" {
		t.Fatalf("expected Hello fallback, got %q", got)
	}
}

// TestCatalogReturnsMissingKey verifies key fallback.
func TestCatalogReturnsMissingKey(t *testing.T) {
	catalog := NewCatalog(Config{DefaultLocale: "es", FallbackLocale: "en"}, testEntries())

	if got := catalog.Default("missing"); got != "missing" {
		t.Fatalf("expected missing key fallback, got %q", got)
	}
}

// TestCatalogReturnsMissingEmpty verifies empty missing mode.
func TestCatalogReturnsMissingEmpty(t *testing.T) {
	catalog := NewCatalog(Config{DefaultLocale: "es", MissingMode: MissingEmpty}, testEntries())

	if got := catalog.Default("missing"); got != "" {
		t.Fatalf("expected empty missing fallback, got %q", got)
	}
}

// TestCatalogInterpolatesParams verifies placeholder replacement.
func TestCatalogInterpolatesParams(t *testing.T) {
	catalog := NewCatalog(Config{DefaultLocale: "es"}, testEntries())

	got := catalog.Default("room.full", Params{"room": "Pixels"})
	if got != "La sala Pixels está llena." {
		t.Fatalf("unexpected interpolation %q", got)
	}
}

// testEntries returns a small translation fixture.
func testEntries() map[Locale]map[Key]string {
	return map[Locale]map[Key]string{
		"es": {
			"hello":     "Hola",
			"room.full": "La sala {room} está llena.",
		},
		"en": {
			"hello": "Hello",
		},
	}
}
