// Package identity owns pet name, appearance, and stat validation.
package identity

import (
	"strings"
	"unicode"
	"unicode/utf8"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

const (
	// NameApproved identifies a valid name.
	NameApproved int32 = iota
	// NameTooLong identifies a name over sixteen runes.
	NameTooLong
	// NameTooShort identifies a name below two runes.
	NameTooShort
	// NameInvalidCharacters identifies unsupported name runes.
	NameInvalidCharacters
	// NameCensored identifies a name rejected by text filtering.
	NameCensored
)

// NormalizeName trims and collapses whitespace without changing letter case.
func NormalizeName(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

// ValidateName returns Nitro's native name validation result code.
func ValidateName(value string) (string, int32) {
	value = NormalizeName(value)
	length := utf8.RuneCountInString(value)
	if length > 16 {
		return value, NameTooLong
	}
	if length < 2 {
		return value, NameTooShort
	}
	for _, current := range value {
		if !unicode.IsLetter(current) && !unicode.IsNumber(current) && current != ' ' && current != '-' && current != '\'' {
			return value, NameInvalidCharacters
		}
	}
	return value, NameApproved
}

// NormalizeColor validates and normalizes a six-digit hexadecimal color.
func NormalizeColor(value string) (string, error) {
	value = strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(value), "#"))
	if len(value) != 6 {
		return "", petrecord.ErrInvalidAppearance
	}
	for _, current := range value {
		if !unicode.IsDigit(current) && (current < 'A' || current > 'F') {
			return "", petrecord.ErrInvalidAppearance
		}
	}
	return value, nil
}
