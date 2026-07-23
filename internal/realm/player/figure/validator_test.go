package figure

import "testing"

// TestValid verifies accepted and rejected bounded figure syntax.
func TestValid(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{value: "hd-180-1.hr-100.ch-210-66.lg-270-82.sh-290-80", valid: true},
		{value: "HD-180-1.hr-100", valid: true},
		{value: "", valid: false},
		{value: "hd", valid: false},
		{value: "hd-0", valid: false},
		{value: "hd-180.hd-181", valid: false},
		{value: "hd-180.", valid: false},
		{value: "h!-180", valid: false},
		{value: "hd-999999999999", valid: false},
	}
	for _, test := range tests {
		if got := Valid(test.value); got != test.valid {
			t.Fatalf("Valid(%q)=%v want %v", test.value, got, test.valid)
		}
	}
}

// BenchmarkValid measures the allocation-free figure syntax hot path.
func BenchmarkValid(benchmark *testing.B) {
	value := "hd-180-1.hr-100.ch-210-66.lg-270-82.sh-290-80"
	benchmark.ReportAllocs()
	for range benchmark.N {
		if !Valid(value) {
			benchmark.Fatal("valid figure rejected")
		}
	}
}
