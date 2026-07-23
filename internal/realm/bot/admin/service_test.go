package admin

import (
	"errors"
	"testing"

	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

// TestValidateKeywordNormalizesSafeWholeWords verifies admin input policy.
func TestValidateKeywordNormalizesSafeWholeWords(t *testing.T) {
	value, err := validateKeyword("  Té-Verde  ", 4)
	if err != nil || value != "té-verde" {
		t.Fatalf("value=%q err=%v", value, err)
	}
	for _, value := range []string{"", "tea pot", "tea!"} {
		if _, err = validateKeyword(value, 4); !errors.Is(err, botrecord.ErrInvalidSkill) {
			t.Fatalf("value=%q err=%v", value, err)
		}
	}
	if _, err = validateKeyword("tea", 0); !errors.Is(err, botrecord.ErrInvalidSkill) {
		t.Fatalf("definition error=%v", err)
	}
}
