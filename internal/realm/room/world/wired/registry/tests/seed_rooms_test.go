package tests

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestDevelopmentLabRoomNamesFitPersistence verifies every QA room satisfies the durable room-name constraint.
func TestDevelopmentLabRoomNamesFitPersistence(t *testing.T) {
	contents, err := os.ReadFile(repositoryPath(t, "internal/realm/room/database/seed/development/0012_rebuild_wired_labs.sql"))
	if err != nil {
		t.Fatal(err)
	}
	expression := regexp.MustCompile(`when 11\d then '([^']+)'`)
	names := statementBody(string(contents), "set name = case id", "end,\n    description")
	matches := expression.FindAllStringSubmatch(names, -1)
	if len(matches) != 6 {
		t.Fatalf("room fixtures=%d, want 6", len(matches))
	}
	for _, match := range matches {
		length := utf8.RuneCountInString(match[1])
		if length < 3 || length > 25 {
			t.Errorf("room name %q has %d characters, want 3..25", match[1], length)
		}
	}
	if !strings.Contains(string(contents), "staff_picked = false") {
		t.Fatal("WIRED QA rooms must not be staff picked")
	}
}
