package currency

import (
	"os"
	"reflect"
	"testing"
)

// TestLoadConfigUsesDefaults verifies default currency settings.
func TestLoadConfigUsesDefaults(t *testing.T) {
	clearCurrencyEnv(t)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.Types != "-1:credits,0:duckets,5:diamonds" {
		t.Fatalf("unexpected currency types %q", config.Types)
	}
	if !reflect.DeepEqual(config.LedgerTypes, []int32{-1}) {
		t.Fatalf("unexpected ledger types %#v", config.LedgerTypes)
	}
}

// TestLoadConfigUsesEnvironment verifies currency setting overrides.
func TestLoadConfigUsesEnvironment(t *testing.T) {
	t.Setenv("PIXELS_CURRENCY_TYPES", "-1:credits,5:diamonds")
	t.Setenv("PIXELS_CURRENCY_LEDGER_TYPES", "-1,5")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.Types != "-1:credits,5:diamonds" || !reflect.DeepEqual(config.LedgerTypes, []int32{-1, 5}) {
		t.Fatalf("unexpected config %#v", config)
	}
}

// clearCurrencyEnv removes currency variables for one test.
func clearCurrencyEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"PIXELS_CURRENCY_TYPES", "PIXELS_CURRENCY_LEDGER_TYPES"} {
		value, found := os.LookupEnv(key)
		_ = os.Unsetenv(key)
		t.Cleanup(func() {
			if found {
				_ = os.Setenv(key, value)
				return
			}
			_ = os.Unsetenv(key)
		})
	}
}
