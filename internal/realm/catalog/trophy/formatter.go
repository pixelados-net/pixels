// Package trophy formats immutable catalog trophy inscriptions.
package trophy

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	// MaxMessageRunes stores the maximum persisted trophy inscription length.
	MaxMessageRunes = 300
	// DateLayout stores the stable date shown by Nitro's trophy widget.
	DateLayout = "02-01-2006"
)

// Filter censors hotel-wide forbidden words.
type Filter interface {
	// Censor applies the current global dictionary.
	Censor(text string) (string, bool)
}

// Formatter creates protocol-compatible trophy extra data.
type Formatter struct {
	// filter applies hotel-wide text policy.
	filter Filter
	// now supplies the inscription time.
	now func() time.Time
}

// New creates a trophy formatter.
func New(filter Filter) *Formatter {
	return &Formatter{filter: filter, now: time.Now}
}

// WithClock replaces the time source for deterministic tests.
func (formatter *Formatter) WithClock(now func() time.Time) *Formatter {
	if now != nil {
		formatter.now = now
	}
	return formatter
}

// Format filters and composes username, date, and buyer text exactly once.
func (formatter *Formatter) Format(username string, text string) string {
	text = strings.Map(func(value rune) rune {
		if value == '\t' || value == '\n' || value == '\r' || value == 0 {
			return ' '
		}
		return value
	}, text)
	text = strings.TrimSpace(text)
	if formatter.filter != nil {
		text, _ = formatter.filter.Censor(text)
	}
	text = truncate(text, MaxMessageRunes)
	username = strings.NewReplacer("\t", " ", "\n", " ", "\r", " ").Replace(strings.TrimSpace(username))
	return username + "\t" + formatter.now().Format(DateLayout) + "\t" + text
}

// truncate returns at most limit Unicode code points.
func truncate(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}
