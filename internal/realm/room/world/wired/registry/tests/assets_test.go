package tests

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// TestArcturusInventoryAudit reproduces the versioned 76/179/113 inventory without a database.
func TestArcturusInventoryAudit(t *testing.T) {
	interactions, rows := baseInteractions(t)
	if rows != 232 || len(interactions) != 179 {
		t.Fatalf("Arcturus WIRED inventory rows=%d types=%d, want 232/179", rows, len(interactions))
	}
	direct := 0
	missing := make([]string, 0, 10)
	for _, descriptor := range registry.CanonicalManifest() {
		if interactions[descriptor.Key] {
			direct++
		} else {
			missing = append(missing, descriptor.Key)
		}
	}
	sort.Strings(missing)
	wantMissing := []string{
		"wf_act_alert", "wf_act_give_respect", "wf_act_match_to_sshot", "wf_act_toggle_to_rnd",
		"wf_cnd_has_handitem", "wf_cnd_not_wearing_b", "wf_cnd_wearing_badge",
		"wf_trg_game_team_lose", "wf_trg_game_team_win", "wf_xtra_or_eval",
	}
	if direct != 66 || len(interactions)-direct != 113 || strings.Join(missing, ",") != strings.Join(wantMissing, ",") {
		t.Fatalf("direct=%d sql-only=%d missing=%v", direct, len(interactions)-direct, missing)
	}
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	if !interactions["wf_act_give_credits"] {
		t.Fatal("corrupt SQL-only fixture disappeared from the upstream audit")
	}
	if _, supported := registered.Resolve("wf_act_give_credits"); supported {
		t.Fatal("corrupt wf_act_give_credits became functional")
	}
}

// TestFunctionalSeedManifest verifies every public asset maps to a registered behavior.
func TestFunctionalSeedManifest(t *testing.T) {
	path := repositoryPath(t, "internal/realm/furniture/database/seed/development/0021_wired_definitions.sql")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	definitions := make(map[string]struct{})
	interactions := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields, valid := tupleFields(scanner.Text())
		if !valid || len(fields) < 18 || !strings.Contains(fields[17], "arcturus-ms3.5.5") {
			continue
		}
		definitions[fields[0]] = struct{}{}
		interaction := unquote(fields[13])
		if _, found := registered.Resolve(interaction); !found {
			t.Fatalf("seeded functional interaction %q is not registered", interaction)
		}
		interactions[interaction] = struct{}{}
	}
	if err = scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(definitions) != 83 || len(interactions) != 69 {
		t.Fatalf("functional assets=%d interaction variants=%d, want 83/69", len(definitions), len(interactions))
	}
}

// baseInteractions reads the versioned WIRED inventory extracted from the pinned Arcturus BaseDB.
func baseInteractions(t *testing.T) (map[string]bool, int) {
	t.Helper()
	file, err := os.Open(repositoryPath(t, "internal/realm/room/world/wired/registry/tests/testdata/arcturus_wired_inventory.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	result := make(map[string]bool)
	rows := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ",")
		if len(fields) != 2 || !strings.HasPrefix(fields[0], "wf_") {
			t.Fatalf("invalid Arcturus WIRED inventory row %q", scanner.Text())
		}
		count, parseErr := strconv.Atoi(fields[1])
		if parseErr != nil || count <= 0 || result[fields[0]] {
			t.Fatalf("invalid Arcturus WIRED inventory row %q", scanner.Text())
		}
		result[fields[0]] = true
		rows += count
	}
	if err = scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return result, rows
}

// tupleFields splits one SQL tuple while preserving quoted commas and escaped quotes.
func tupleFields(line string) ([]string, bool) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "(") {
		return nil, false
	}
	line = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(line, ";"), ","), ")")
	line = strings.TrimPrefix(line, "(")
	fields := make([]string, 0, 24)
	start, quoted := 0, false
	for index := 0; index < len(line); index++ {
		switch line[index] {
		case '\'':
			if index > 0 && line[index-1] == '\\' {
				continue
			}
			if quoted && index+1 < len(line) && line[index+1] == '\'' {
				index++
				continue
			}
			quoted = !quoted
		case ',':
			if !quoted {
				fields = append(fields, strings.TrimSpace(line[start:index]))
				start = index + 1
			}
		}
	}
	fields = append(fields, strings.TrimSpace(line[start:]))
	return fields, !quoted
}

// unquote removes one SQL string delimiter pair and unescapes doubled quotes.
func unquote(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
		value = value[1 : len(value)-1]
	}
	return strings.ReplaceAll(value, "''", "'")
}

// repositoryPath resolves a repository-relative fixture from this source file.
func repositoryPath(t *testing.T, relative string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../../", relative))
}
