package trade

import (
	"testing"
	"time"
)

// TestLoadConfigNormalizesUnsafeValues verifies bounded trade runtime policy.
func TestLoadConfigNormalizesUnsafeValues(t *testing.T) {
	t.Setenv("PIXELS_TRADE_START_THROTTLE", "0s")
	t.Setenv("PIXELS_TRADE_MAX_ITEMS", "1000")
	config := LoadConfig()
	if config.StartThrottle != 10*time.Second || config.MaximumItems != 100 {
		t.Fatalf("config=%#v", config)
	}
}
