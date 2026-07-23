package i18n

import (
	"os"
	"testing"
)

// clearEnv removes environment variables for a test and restores them afterward.
func clearEnv(t *testing.T, keys ...string) {
	t.Helper()

	values := make(map[string]string, len(keys))
	present := make(map[string]bool, len(keys))
	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		values[key] = value
		present[key] = ok
		_ = os.Unsetenv(key)
	}

	t.Cleanup(func() {
		for _, key := range keys {
			if present[key] {
				_ = os.Setenv(key, values[key])
				continue
			}

			_ = os.Unsetenv(key)
		}
	})
}
