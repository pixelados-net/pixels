// Package textfilter matches and censors whole Unicode words.
package textfilter

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Contains reports whether text contains a case-insensitive whole-word match.
func Contains(text string, words []string) bool {
	for offset := 0; offset < len(text); {
		start, end := nextToken(text, offset)
		if start == end {
			break
		}
		if matches(text[start:end], words) {
			return true
		}
		offset = end
	}

	return false
}

// Censor replaces matching whole words with stars while preserving rune count.
func Censor(text string, words []string) (string, bool) {
	var builder strings.Builder
	last := 0
	matched := false
	for offset := 0; offset < len(text); {
		start, end := nextToken(text, offset)
		if start == end {
			break
		}
		if matches(text[start:end], words) {
			if !matched {
				builder.Grow(len(text))
			}
			builder.WriteString(text[last:start])
			for range text[start:end] {
				builder.WriteByte('*')
			}
			last = end
			matched = true
		}
		offset = end
	}
	if !matched {
		return text, false
	}
	builder.WriteString(text[last:])

	return builder.String(), true
}

// matches reports whether one token matches an immutable word snapshot.
func matches(token string, words []string) bool {
	for _, word := range words {
		if strings.EqualFold(token, word) {
			return true
		}
	}

	return false
}

// nextToken returns the next letter-or-number token bounds without allocation.
func nextToken(text string, offset int) (int, int) {
	start := offset
	for start < len(text) {
		r, size := utf8.DecodeRuneInString(text[start:])
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			break
		}
		start += size
	}
	end := start
	for end < len(text) {
		r, size := utf8.DecodeRuneInString(text[end:])
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			break
		}
		end += size
	}

	return start, end
}
