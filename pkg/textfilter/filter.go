// Package textfilter matches and censors normalized Unicode text.
package textfilter

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Contains reports whether text contains a compiled pattern.
func (matcher *Matcher) Contains(text string) bool {
	if matcher == nil || len(matcher.nodes) <= 1 {
		return false
	}
	state := 0
	for _, value := range text {
		if !unicode.IsLetter(value) && !unicode.IsNumber(value) {
			continue
		}
		state = matcher.next(state, unicode.ToLower(value))
		if len(matcher.nodes[state].outputs) != 0 {
			return true
		}
	}

	return false
}

// Censor replaces matching letters and numbers while preserving separators.
func (matcher *Matcher) Censor(text string) (string, bool) {
	if !matcher.Contains(text) {
		return text, false
	}
	masked := make([]bool, len(text))
	positions := make([]int, matcher.maxPattern)
	state := 0
	index := 0
	for offset, value := range text {
		if !unicode.IsLetter(value) && !unicode.IsNumber(value) {
			continue
		}
		positions[index%matcher.maxPattern] = offset
		state = matcher.next(state, unicode.ToLower(value))
		for _, length := range matcher.nodes[state].outputs {
			for matched := index - length + 1; matched <= index; matched++ {
				masked[positions[matched%matcher.maxPattern]] = true
			}
		}
		index++
	}
	var builder strings.Builder
	builder.Grow(len(text))
	for offset := 0; offset < len(text); {
		value, size := utf8.DecodeRuneInString(text[offset:])
		if masked[offset] && (unicode.IsLetter(value) || unicode.IsNumber(value)) {
			builder.WriteByte('*')
		} else {
			builder.WriteString(text[offset : offset+size])
		}
		offset += size
	}

	return builder.String(), true
}

// Contains reports whether text contains any supplied pattern.
func Contains(text string, words []string) bool {
	return Compile(words).Contains(text)
}

// Censor replaces matches from a newly compiled pattern set.
func Censor(text string, words []string) (string, bool) {
	return Compile(words).Censor(text)
}
