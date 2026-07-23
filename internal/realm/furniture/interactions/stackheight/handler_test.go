package stackheight

import (
	"testing"

	instack "github.com/niflaot/pixels/networking/inbound/furniture/stackheight"
)

// TestNormalizedHeight verifies automatic, exact, and rejected centimeter values.
func TestNormalizedHeight(t *testing.T) {
	automatic, err := normalizedHeight(instack.AutoHeight)
	if err != nil || automatic != nil {
		t.Fatalf("automatic=%v err=%v", automatic, err)
	}
	for _, value := range []int32{0, 1, MaxHeightCM} {
		height, err := normalizedHeight(value)
		if err != nil || height == nil || *height != value {
			t.Fatalf("value=%d height=%v err=%v", value, height, err)
		}
	}
	for _, value := range []int32{-1, MaxHeightCM + 1} {
		if _, err := normalizedHeight(value); err == nil {
			t.Fatalf("value=%d expected error", value)
		}
	}
}

// BenchmarkNormalizedHeight measures the slider validation path.
func BenchmarkNormalizedHeight(b *testing.B) {
	for range b.N {
		_, _ = normalizedHeight(123)
	}
}
