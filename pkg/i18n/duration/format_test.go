package duration

import (
	"testing"
	"time"

	"github.com/niflaot/pixels/pkg/i18n"
)

// formatCase stores one duration formatting expectation.
type formatCase struct {
	// name identifies the test case.
	name string
	// value stores the duration to format.
	value time.Duration
	// expected stores the expected Spanish output.
	expected string
}

// TestFormatSelectsLocalizedUnits verifies boundaries, plural forms, and rounding.
func TestFormatSelectsLocalizedUnits(t *testing.T) {
	catalog := testCatalog()
	cases := []formatCase{
		{name: "minimum", value: 0, expected: "1 segundo"},
		{name: "seconds", value: 59 * time.Second, expected: "59 segundos"},
		{name: "minute", value: time.Minute, expected: "1 minuto"},
		{name: "rounded minutes", value: 61 * time.Second, expected: "2 minutos"},
		{name: "hour", value: time.Hour, expected: "1 hora"},
		{name: "rounded hours", value: 61 * time.Minute, expected: "2 horas"},
		{name: "day", value: 24 * time.Hour, expected: "1 día"},
		{name: "rounded days", value: 25 * time.Hour, expected: "2 días"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := Default(catalog, test.value); got != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, got)
			}
		})
	}
	if got := Format(catalog, "en", time.Hour); got != "1 hour" {
		t.Fatalf("expected English locale, got %q", got)
	}
}

// TestDefaultParamsBuildsInterpolation verifies parent-message parameters.
func TestDefaultParamsBuildsInterpolation(t *testing.T) {
	params := DefaultParams(testCatalog(), 90*time.Second)
	if params["duration"] != "2 minutos" {
		t.Fatalf("unexpected duration parameter %#v", params)
	}
}

// TestNilTranslatorReturnsEmpty verifies optional translator behavior.
func TestNilTranslatorReturnsEmpty(t *testing.T) {
	if got := Default(nil, time.Minute); got != "" {
		t.Fatalf("expected empty text, got %q", got)
	}
}

// BenchmarkDefault measures localized duration formatting cost.
func BenchmarkDefault(b *testing.B) {
	catalog := testCatalog()
	b.ReportAllocs()
	for b.Loop() {
		_ = Default(catalog, 10*time.Minute)
	}
}

// testCatalog creates a bilingual duration fixture.
func testCatalog() *i18n.Catalog {
	return i18n.NewCatalog(i18n.Config{DefaultLocale: "es", FallbackLocale: "en"}, map[i18n.Locale]map[i18n.Key]string{
		"es": {
			secondOne: "1 segundo", secondOther: "{count} segundos",
			minuteOne: "1 minuto", minuteOther: "{count} minutos",
			hourOne: "1 hora", hourOther: "{count} horas",
			dayOne: "1 día", dayOther: "{count} días",
		},
		"en": {hourOne: "1 hour"},
	})
}
