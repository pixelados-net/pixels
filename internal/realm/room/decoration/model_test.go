package decoration

import "testing"

// TestPostItColorSeparatesVisualStateFromText verifies Nitro receives only its supported color token.
func TestPostItColorSeparatesVisualStateFromText(t *testing.T) {
	for _, test := range []struct {
		value string
		color string
	}{
		{value: "9CCEFF blue note", color: "9CCEFF"},
		{value: "FF9CFF pink note", color: "FF9CFF"},
		{value: "9CFF9C green note", color: "9CFF9C"},
		{value: DefaultPostItData, color: DefaultPostItColor},
		{value: "0", color: DefaultPostItColor},
		{value: "000000 invalid", color: DefaultPostItColor},
	} {
		if color := PostItColor(test.value); color != test.color {
			t.Fatalf("PostItColor(%q) = %q, want %q", test.value, color, test.color)
		}
	}
}

// BenchmarkPostItColor measures the room-entry post-it projection hot path.
func BenchmarkPostItColor(b *testing.B) {
	const value = "9CFF9C visible text"
	b.ReportAllocs()
	for b.Loop() {
		if PostItColor(value) != "9CFF9C" {
			b.Fatal("unexpected projected color")
		}
	}
}
