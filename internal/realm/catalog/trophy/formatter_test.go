package trophy

import (
	"strings"
	"testing"
	"time"
)

// filterForTest replaces one fixture word.
type filterForTest struct{}

// Censor applies one deterministic fixture replacement.
func (filterForTest) Censor(text string) (string, bool) {
	return strings.ReplaceAll(text, "bad", "***"), strings.Contains(text, "bad")
}

// TestFormatCreatesSafeProtocolData verifies filtering, controls, date, and truncation.
func TestFormatCreatesSafeProtocolData(t *testing.T) {
	t.Parallel()
	formatter := New(filterForTest{}).WithClock(func() time.Time {
		return time.Date(2026, time.July, 14, 12, 0, 0, 0, time.UTC)
	})
	message := "bad\tline\n" + strings.Repeat("á", MaxMessageRunes)
	formatted := formatter.Format(" demo\n", message)
	parts := strings.Split(formatted, "\t")
	if len(parts) != 3 || parts[0] != "demo" || parts[1] != "14-07-2026" || strings.ContainsAny(parts[2], "\t\n\r") || !strings.HasPrefix(parts[2], "*** line ") {
		t.Fatalf("unexpected trophy data %q", formatted)
	}
	if len([]rune(parts[2])) != MaxMessageRunes {
		t.Fatalf("got %d message runes", len([]rune(parts[2])))
	}
}

// BenchmarkFormat measures the purchase-time inscription path.
func BenchmarkFormat(b *testing.B) {
	formatter := New(filterForTest{})
	b.ReportAllocs()
	for b.Loop() {
		_ = formatter.Format("demo", "A clean trophy inscription")
	}
}
