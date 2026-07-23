package textfilter

import "testing"

// TestFilter verifies normalized substring and obfuscation censorship.
func TestFilter(t *testing.T) {
	matcher := Compile([]string{"bad", "niño", "chancleta"})
	if !matcher.Contains("BAD idea") || !matcher.Contains("badge") || !matcher.Contains("chan cleta") {
		t.Fatal("unexpected normalized match")
	}
	text, changed := matcher.Censor("bad badge niño chan cleta chancletacion")
	if !changed || text != "*** ***ge **** **** ***** *********cion" {
		t.Fatalf("text=%q changed=%v", text, changed)
	}
	text, changed = matcher.Censor("clean")
	if changed || text != "clean" {
		t.Fatalf("text=%q changed=%v", text, changed)
	}
}

// TestFailureTransitions verifies overlapping suffix and fallback matches.
func TestFailureTransitions(t *testing.T) {
	matcher := Compile([]string{"he", "she", "hers", "his", "", "she"})
	if !matcher.Contains("ushers") || !matcher.Contains("H-I-S") || matcher.Contains("plain") {
		t.Fatal("unexpected automaton transition result")
	}
	text, changed := matcher.Censor("ushers")
	if !changed || text != "u*****" {
		t.Fatalf("text=%q changed=%v", text, changed)
	}
}

// BenchmarkCensor measures clean and matching filter paths.
func BenchmarkCensor(b *testing.B) {
	matcher := Compile([]string{"blocked", "another", "niño"})
	for _, text := range []string{"a normal room message", "a blocked message for another niño"} {
		b.Run(text, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_, _ = matcher.Censor(text)
			}
		})
	}
}
