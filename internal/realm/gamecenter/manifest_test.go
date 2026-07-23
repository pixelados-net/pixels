package gamecenter

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

var gamesHeaderPattern = regexp.MustCompile(`Header\s+uint16\s*=\s*([0-9]+)`)

// gamesPacket identifies one direction-specific Games protocol packet.
type gamesPacket struct {
	// direction stores inbound or outbound.
	direction string
	// header stores the wire identifier.
	header uint16
}

// TestGamesPacketManifest verifies the exact 49-packet audited Games surface including PLAYING_GAME.
func TestGamesPacketManifest(t *testing.T) {
	root := gamesRepositoryRoot(t)
	packets := scanGamesPackets(t, root)
	expected := expectedGamesPackets()
	if len(packets) != 49 {
		t.Fatalf("Games surface has %d direction/header pairs, want 49", len(packets))
	}
	seen := make(map[gamesPacket]struct{}, len(packets))
	for _, packet := range packets {
		if _, duplicate := seen[packet]; duplicate {
			t.Fatalf("duplicate %s:%d", packet.direction, packet.header)
		}
		seen[packet] = struct{}{}
	}
	for _, packet := range expected {
		if _, found := seen[packet]; !found {
			t.Errorf("missing %s:%d", packet.direction, packet.header)
		}
	}
}

// scanGamesPackets reads Game Center, DB poll, infobus, and playing-game packages.
func scanGamesPackets(t testing.TB, root string) []gamesPacket {
	t.Helper()
	paths := []string{"networking/inbound/gamecenter", "networking/outbound/gamecenter"}
	packets := make([]gamesPacket, 0, 49)
	for _, relative := range paths {
		err := filepath.WalkDir(filepath.Join(root, relative), func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() || entry.Name() != "packet.go" {
				return walkErr
			}
			packets = append(packets, readGamesPacket(t, root, path))
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	extras := []string{
		"networking/inbound/progression/poll/start/packet.go", "networking/inbound/progression/poll/reject/packet.go", "networking/inbound/progression/poll/answer/packet.go", "networking/inbound/progression/poll/infobusvote/packet.go",
		"networking/outbound/progression/poll/contents/packet.go", "networking/outbound/progression/poll/error/packet.go", "networking/outbound/progression/poll/infobusresult/packet.go", "networking/outbound/progression/poll/infobusstart/packet.go", "networking/outbound/progression/poll/offer/packet.go", "networking/outbound/room/games/playing/packet.go",
	}
	for _, relative := range extras {
		packets = append(packets, readGamesPacket(t, root, filepath.Join(root, relative)))
	}
	return packets
}

// readGamesPacket validates one packet implementation and its golden test sibling.
func readGamesPacket(t testing.TB, root string, path string) gamesPacket {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	match := gamesHeaderPattern.FindSubmatch(data)
	if len(match) != 2 {
		t.Fatalf("missing Header in %s", path)
	}
	value, err := strconv.ParseUint(string(match[1]), 10, 16)
	if err != nil {
		t.Fatal(err)
	}
	testPath := strings.TrimSuffix(path, ".go") + "_test.go"
	if _, err = os.Stat(testPath); err != nil {
		t.Fatalf("missing golden test %s", testPath)
	}
	direction := "inbound"
	relative, _ := filepath.Rel(root, path)
	if strings.Contains(filepath.ToSlash(relative), "networking/outbound/") {
		direction = "outbound"
	}
	return gamesPacket{direction: direction, header: uint16(value)}
}

// gamesRepositoryRoot resolves the checkout root from this test source.
func gamesRepositoryRoot(t testing.TB) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}

// expectedGamesPackets returns the exact audited direction/header set.
func expectedGamesPackets() []gamesPacket {
	inbound := []uint16{11, 109, 741, 1054, 1232, 1445, 1458, 1598, 1773, 2384, 2415, 2502, 2565, 2914, 3171, 3196, 3207, 3259, 3505, 3802, 6200}
	outbound := []uint16{222, 448, 662, 872, 904, 1477, 1715, 1730, 2142, 2196, 2246, 2260, 2270, 2624, 2641, 2893, 2997, 3035, 3097, 3138, 3191, 3512, 3560, 3654, 3785, 3805, 5200, 5201}
	packets := make([]gamesPacket, 0, len(inbound)+len(outbound))
	for _, header := range inbound {
		packets = append(packets, gamesPacket{direction: "inbound", header: header})
	}
	for _, header := range outbound {
		packets = append(packets, gamesPacket{direction: "outbound", header: header})
	}
	return packets
}
