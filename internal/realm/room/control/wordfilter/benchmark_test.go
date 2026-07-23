package wordfilter

import (
	"testing"

	"github.com/niflaot/pixels/pkg/textfilter"
)

// BenchmarkCensor measures the room-filter hot path with and without matches.
func BenchmarkCensor(b *testing.B) {
	words := []string{"blocked", "another", "niño"}
	matcher := textfilter.Compile(words)
	for _, test := range []struct {
		name string
		text string
	}{
		{name: "clean", text: "a normal room message without filtered content"},
		{name: "matched", text: "a blocked room message for another niño"},
	} {
		b.Run(test.name, func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_, _ = matcher.Censor(test.text)
			}
		})
	}
}
