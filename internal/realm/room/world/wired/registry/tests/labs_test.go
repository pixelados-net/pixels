package tests

import (
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// labItem stores the static facts required to audit one regenerated QA item.
type labItem struct {
	definitionID int64
	roomID       int64
	x            int
	y            int
}

// TestRegeneratedLabsUseExecutableStacks verifies behaviors are connected instead of merely displayed.
func TestRegeneratedLabsUseExecutableStacks(t *testing.T) {
	contents := rebuiltLabSQL(t)
	definitions := seededDefinitions(t)
	items := labItems(t, contents)
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	type families struct {
		trigger   bool
		condition bool
		effect    bool
	}
	stacks := make(map[[3]int64]families)
	behaviors := make(map[string][3]int64)
	for itemID, item := range items {
		descriptor, found := registered.Resolve(definitions[item.definitionID])
		if !found {
			continue
		}
		key := [3]int64{item.roomID, int64(item.x), int64(item.y)}
		stack := stacks[key]
		switch descriptor.Family {
		case registry.FamilyTrigger:
			stack.trigger = true
		case registry.FamilyCondition:
			stack.condition = true
		case registry.FamilyEffect:
			stack.effect = true
		}
		stacks[key] = stack
		behaviors[descriptor.Key] = key
		_ = itemID
	}
	for _, descriptor := range append(registry.CanonicalManifest(), registry.CompatibilityManifest()...) {
		key, found := behaviors[descriptor.Key]
		if !found || descriptor.Key == "wf_blob" || descriptor.Key == "wf_highscore" {
			continue
		}
		stack := stacks[key]
		switch descriptor.Family {
		case registry.FamilyTrigger:
			if !stack.effect {
				t.Errorf("trigger %s has no observable effect", descriptor.Key)
			}
		case registry.FamilyCondition:
			if !stack.trigger || !stack.effect {
				t.Errorf("condition %s is not in an executable trigger/effect stack", descriptor.Key)
			}
		case registry.FamilyEffect:
			if !stack.trigger {
				t.Errorf("effect %s has no trigger", descriptor.Key)
			}
		}
	}
}

// TestRegeneratedLabsUseWalkableArrivalTargets guards the walk-on, walk-off, and bot-arrival fixtures.
func TestRegeneratedLabsUseWalkableArrivalTargets(t *testing.T) {
	contents := rebuiltLabSQL(t)
	items := labItems(t, contents)
	targets := selectedTargets(contents)
	for _, wiredID := range []int64{421020, 421200, 422010, 422020, 424070} {
		targetID, found := targets[wiredID]
		if !found {
			t.Errorf("wired item %d has no selected target", wiredID)
			continue
		}
		if target := items[targetID]; target.definitionID != 28 {
			t.Errorf("wired item %d target definition=%d, want walkable pressure plate 28", wiredID, target.definitionID)
		}
	}
}

// TestRegeneratedLabSayCommandsDoNotOverlap guards substring-matched QA triggers.
func TestRegeneratedLabSayCommandsDoNotOverlap(t *testing.T) {
	contents := rebuiltLabSQL(t)
	parameters := wiredStringParameters(contents)
	correction, err := os.ReadFile(repositoryPath(t, "internal/realm/furniture/database/seed/development/0027_fix_wired_lab_say_commands.sql"))
	if err != nil {
		t.Fatal(err)
	}
	for itemID, value := range wiredStringParameters(string(correction)) {
		parameters[itemID] = value
	}
	definitions := seededDefinitions(t)
	commands := make(map[int64][]sayCommand)
	for itemID, item := range labItems(t, contents) {
		if definitions[item.definitionID] != "wf_trg_says_something" {
			continue
		}
		command := strings.ToLower(strings.TrimSpace(parameters[itemID]))
		if command == "" {
			t.Errorf("say trigger %d has no command", itemID)
			continue
		}
		commands[item.roomID] = append(commands[item.roomID], sayCommand{itemID: itemID, value: command})
	}
	for roomID, roomCommands := range commands {
		sort.Slice(roomCommands, func(left int, right int) bool {
			return roomCommands[left].itemID < roomCommands[right].itemID
		})
		for left := 0; left < len(roomCommands); left++ {
			for right := left + 1; right < len(roomCommands); right++ {
				first := roomCommands[left]
				second := roomCommands[right]
				if strings.Contains(first.value, second.value) || strings.Contains(second.value, first.value) {
					t.Errorf("room %d say commands overlap: %d=%q and %d=%q", roomID, first.itemID, first.value, second.itemID, second.value)
				}
			}
		}
	}
}

// TestRegeneratedCollisionLabUsesDedicatedTarget guards the repeatable furniture-to-unit fixture.
func TestRegeneratedCollisionLabUsesDedicatedTarget(t *testing.T) {
	contents, err := os.ReadFile(repositoryPath(t, "internal/realm/furniture/database/seed/development/0030_regenerate_wired_movement_lab.sql"))
	if err != nil {
		t.Fatal(err)
	}
	items := labItems(t, string(contents))
	targets := selectedTargets(string(contents))
	parameters := wiredStringParameters(string(contents))
	target := items[429203]
	if target.definitionID != 29 || target.roomID != 112 || target.x != 10 || target.y != 8 {
		t.Fatalf("collision target=%+v, want Floor Switch in room 112 at (10,8)", target)
	}
	if targets[422151] != 429203 {
		t.Fatalf("collision movement target=%d, want 429203", targets[422151])
	}
	if parameters[422150] != "collision" {
		t.Fatalf("collision command=%q", parameters[422150])
	}
}

// TestRegeneratedSnapshotLabUsesDedicatedTarget guards the positive and negative snapshot fixtures.
func TestRegeneratedSnapshotLabUsesDedicatedTarget(t *testing.T) {
	contents := rebuiltLabSQL(t)
	items := labItems(t, contents)
	targets := selectedTargets(contents)
	correction, err := os.ReadFile(repositoryPath(t, "internal/realm/furniture/database/seed/development/0029_clarify_wired_snapshot_lab.sql"))
	if err != nil {
		t.Fatal(err)
	}
	const targetID int64 = 429106
	for _, wiredID := range []int64{421081, 421091} {
		if targets[wiredID] != targetID {
			t.Errorf("snapshot condition %d target=%d, want %d", wiredID, targets[wiredID], targetID)
		}
	}
	target := items[targetID]
	if target.roomID != 111 || target.x != 9 || target.y != 12 {
		t.Errorf("snapshot target=%+v, want room 111 rightmost item at (9,12)", target)
	}
	for _, wiredID := range []int64{421060, 421061, 421070, 421071} {
		if targets[wiredID] == targetID {
			t.Errorf("state-changed fixture %d must not reuse snapshot target %d", wiredID, targetID)
		}
	}
	clarification := string(correction)
	if !strings.Contains(clarification, "set definition_id=27") || !strings.Contains(clarification, "where id=429106 and room_id=111") {
		t.Error("snapshot correction must give item 429106 its dedicated Color Wheel definition")
	}
	if !strings.Contains(clarification, "RIGHTMOST Color Wheel (ID 429106)") {
		t.Error("snapshot correction must explain the exact target in the room-entry guide")
	}
}

// sayCommand identifies one configured says-something QA trigger.
type sayCommand struct {
	// itemID identifies the configured WIRED trigger.
	itemID int64
	// value stores the normalized speech keyword.
	value string
}

// rebuiltLabSQL reads the authoritative replacement fixture.
func rebuiltLabSQL(t *testing.T) string {
	t.Helper()
	contents, err := os.ReadFile(repositoryPath(t, "internal/realm/furniture/database/seed/development/0026_rebuild_wired_labs.sql"))
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

// labItems parses item identity, room, and stack coordinates from the deterministic seed.
func labItems(t *testing.T, contents string) map[int64]labItem {
	t.Helper()
	expression := regexp.MustCompile(`\((\d+),(\d+),1,(11[0-5]),(\d+),(\d+),[0-9.]+,`)
	result := make(map[int64]labItem)
	for _, match := range expression.FindAllStringSubmatch(statementBody(contents, "insert into furniture_items", "on conflict(id)"), -1) {
		values := make([]int64, 5)
		for index := range values {
			parsed, err := strconv.ParseInt(match[index+1], 10, 64)
			if err != nil {
				t.Fatalf("parse lab tuple %v: %v", match, err)
			}
			values[index] = parsed
		}
		result[values[0]] = labItem{definitionID: values[1], roomID: values[2], x: int(values[3]), y: int(values[4])}
	}
	return result
}

// selectedTargets parses the first target of each WIRED fixture.
func selectedTargets(contents string) map[int64]int64 {
	statement := statementBody(contents, "insert into room_wired_selected_items", "on conflict(wired_item_id")
	expression := regexp.MustCompile(`\((\d+),(\d+),0,`)
	result := make(map[int64]int64)
	for _, match := range expression.FindAllStringSubmatch(statement, -1) {
		wiredID, wiredErr := strconv.ParseInt(match[1], 10, 64)
		targetID, targetErr := strconv.ParseInt(match[2], 10, 64)
		if wiredErr == nil && targetErr == nil {
			result[wiredID] = targetID
		}
	}
	return result
}

// wiredStringParameters parses normalized WIRED string parameters from a seed.
func wiredStringParameters(contents string) map[int64]string {
	expression := regexp.MustCompile(`\((\d+),'[^']*','([^']*)',\d+,\d+\)`)
	result := make(map[int64]string)
	for _, match := range expression.FindAllStringSubmatch(contents, -1) {
		itemID, err := strconv.ParseInt(match[1], 10, 64)
		if err == nil {
			result[itemID] = match[2]
		}
	}
	return result
}
