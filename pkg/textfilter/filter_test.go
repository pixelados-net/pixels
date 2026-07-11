package textfilter

import "testing"

// TestFilter verifies whole-word matching and Unicode-preserving censorship.
func TestFilter(t *testing.T) {
	words := []string{"bad", "niño"}
	if !Contains("BAD idea", words) || Contains("badge", words) {
		t.Fatal("unexpected whole-word match")
	}
	text, changed := Censor("bad badge niño", words)
	if !changed || text != "*** badge ****" {
		t.Fatalf("text=%q changed=%v", text, changed)
	}
	text, changed = Censor("clean", words)
	if changed || text != "clean" {
		t.Fatalf("text=%q changed=%v", text, changed)
	}
}

// BenchmarkCensor measures clean and matching filter paths.
func BenchmarkCensor(b *testing.B) {
	words := []string{"blocked", "another", "niño"}
	for _, text := range []string{"a normal room message", "a blocked message for another niño"} {
		b.Run(text, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_, _ = Censor(text, words)
			}
		})
	}
}
