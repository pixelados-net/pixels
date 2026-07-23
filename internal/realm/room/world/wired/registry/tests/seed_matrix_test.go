package tests

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// TestDevelopmentLabsPlaceAndConfigureManifest verifies all canonical and compatibility behaviors have QA fixtures.
func TestDevelopmentLabsPlaceAndConfigureManifest(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	definitions := seededDefinitions(t)
	placed := make(map[string]struct{})
	configured := make(map[string]struct{})
	for _, relative := range []string{
		"internal/realm/furniture/database/seed/development/0026_rebuild_wired_labs.sql",
		"internal/realm/furniture/database/seed/development/0038_room_games.sql",
	} {
		contents, readErr := os.ReadFile(repositoryPath(t, relative))
		if readErr != nil {
			t.Fatal(readErr)
		}
		for itemID, definitionID := range placedItems(t, string(contents)) {
			interaction := definitions[definitionID]
			descriptor, found := registered.Resolve(interaction)
			if !found {
				continue
			}
			placed[descriptor.Key] = struct{}{}
			if configuredItems(string(contents))[itemID] {
				configured[descriptor.Key] = struct{}{}
			}
		}
	}
	want := append(registry.CanonicalManifest(), registry.CompatibilityManifest()...)
	for _, descriptor := range want {
		if _, found := placed[descriptor.Key]; !found {
			t.Errorf("behavior %s has no placed QA fixture", descriptor.Key)
		}
		if requiresSettings(descriptor) {
			if _, found := configured[descriptor.Key]; !found {
				t.Errorf("behavior %s has no configured QA fixture", descriptor.Key)
			}
		}
	}
	if len(placed) != len(want) {
		t.Fatalf("placed behavior keys=%d, want %d", len(placed), len(want))
	}
}

// seededDefinitions maps development furniture definition ids to interactions.
func seededDefinitions(t *testing.T) map[int64]string {
	t.Helper()
	result := make(map[int64]string)
	for _, relative := range []string{
		"internal/realm/furniture/database/seed/development/0021_wired_definitions.sql",
		"internal/realm/furniture/database/seed/development/0023_wired_compatibility.sql",
		"internal/realm/furniture/database/seed/development/0038_room_games.sql",
	} {
		contents, err := os.ReadFile(repositoryPath(t, relative))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(contents), "\n") {
			fields, valid := tupleFields(line)
			if !valid || len(fields) < 14 {
				continue
			}
			id, parseErr := strconv.ParseInt(fields[0], 10, 64)
			if parseErr == nil {
				result[id] = unquote(fields[13])
			}
		}
	}
	return result
}

// placedItems returns QA furniture ids and their definition ids.
func placedItems(t *testing.T, contents string) map[int64]int64 {
	t.Helper()
	expression := regexp.MustCompile(`\((\d+),(\d+),1,11[0-5],`)
	result := make(map[int64]int64)
	for _, match := range expression.FindAllStringSubmatch(contents, -1) {
		itemID, itemErr := strconv.ParseInt(match[1], 10, 64)
		definitionID, definitionErr := strconv.ParseInt(match[2], 10, 64)
		if itemErr != nil || definitionErr != nil {
			t.Fatalf("invalid placed item tuple %v", match)
		}
		result[itemID] = definitionID
	}
	return result
}

// configuredItems returns item ids from one normalized settings statement.
func configuredItems(contents string) map[int64]bool {
	statement := statementBody(contents, "insert into room_wired_settings", "on conflict(item_id)")
	expression := regexp.MustCompile(`\((\d+),`)
	result := make(map[int64]bool)
	for _, match := range expression.FindAllStringSubmatch(statement, -1) {
		if itemID, err := strconv.ParseInt(match[1], 10, 64); err == nil {
			result[itemID] = true
		}
	}
	return result
}

// statementBody returns SQL between one insertion prefix and conflict clause.
func statementBody(contents string, start string, end string) string {
	startIndex := strings.Index(contents, start)
	if startIndex < 0 {
		return ""
	}
	contents = contents[startIndex:]
	endIndex := strings.Index(contents, end)
	if endIndex < 0 {
		return contents
	}
	return contents[:endIndex]
}

// requiresSettings reports whether one QA behavior compiles persisted configuration.
func requiresSettings(descriptor registry.Descriptor) bool {
	return descriptor.Family != registry.FamilyHighscore && descriptor.Key != "wf_blob"
}
