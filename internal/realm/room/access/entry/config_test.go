package entry

import (
	"os"
	"testing"
	"time"
)

// TestLoadConfigUsesDefaults verifies closed-room entry defaults.
func TestLoadConfigUsesDefaults(t *testing.T) {
	keys := []string{"PIXELS_ROOM_ENTRY_HANGOUT_TIMEOUT", "PIXELS_ROOM_ENTRY_MAX_PASSWORD_ATTEMPTS", "PIXELS_ROOM_ENTRY_ATTEMPT_WINDOW", "PIXELS_ROOM_ENTRY_LOCKOUT_SECONDS", "PIXELS_ROOM_ENTRY_PASSWORD_COST", "PIXELS_ROOM_ENTRY_TRUSTED_TTL"}
	clearEntryEnv(t, keys...)
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	config = config.Normalize()
	if config.HangoutTimeout != 5*time.Minute || config.MaxPasswordAttempts != 5 || config.LockoutDuration() != 10*time.Minute {
		t.Fatalf("unexpected defaults %#v", config)
	}
}

// clearEntryEnv removes entry variables and restores them after the test.
func clearEntryEnv(t *testing.T, keys ...string) {
	t.Helper()
	values := make(map[string]string, len(keys))
	present := make(map[string]bool, len(keys))
	for _, key := range keys {
		values[key], present[key] = os.LookupEnv(key)
		_ = os.Unsetenv(key)
	}
	t.Cleanup(func() {
		for _, key := range keys {
			if present[key] {
				_ = os.Setenv(key, values[key])
			} else {
				_ = os.Unsetenv(key)
			}
		}
	})
}

// TestLoadConfigUsesEnvironment verifies closed-room entry overrides.
func TestLoadConfigUsesEnvironment(t *testing.T) {
	t.Setenv("PIXELS_ROOM_ENTRY_HANGOUT_TIMEOUT", "2m")
	t.Setenv("PIXELS_ROOM_ENTRY_MAX_PASSWORD_ATTEMPTS", "3")
	t.Setenv("PIXELS_ROOM_ENTRY_LOCKOUT_SECONDS", "240")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.HangoutTimeout != 2*time.Minute || config.MaxPasswordAttempts != 3 || config.LockoutDuration() != 4*time.Minute {
		t.Fatalf("unexpected config %#v", config)
	}
}

// TestNormalizeClampsLockoutSeconds verifies duration overflow protection.
func TestNormalizeClampsLockoutSeconds(t *testing.T) {
	config := Config{LockoutSeconds: 1<<63 - 1}.Normalize()
	if config.LockoutSeconds != maxLockoutSeconds || config.LockoutDuration() <= 0 {
		t.Fatalf("unexpected clamped duration %#v", config)
	}
}
