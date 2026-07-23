package progression

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// manifestHeaderPattern finds the repository's packet header declaration.
var manifestHeaderPattern = regexp.MustCompile(`Header\s+uint16\s*=\s*([0-9]+)`)

// packetIdentity identifies one direction-specific protocol header.
type packetIdentity struct {
	// direction stores inbound or outbound.
	direction string
	// header stores the unsigned wire identifier.
	header uint16
}

// TestProgressionPacketManifest verifies the complete audited progression surface remains present.
func TestProgressionPacketManifest(t *testing.T) {
	root := repositoryRoot(t)
	identities := progressionPackets(t, root)
	if len(identities) != 63 {
		t.Fatalf("progression packet tree has %d pairs, want 63", len(identities))
	}
	adjacent := []string{
		"networking/inbound/user/achievement/list/packet.go",
		"networking/outbound/user/achievement/list/packet.go",
		"networking/inbound/navigator/browse/eventcats/packet.go",
		"networking/outbound/camera/competitionstatus/packet.go",
		"networking/inbound/messenger/friend/questcomplete/packet.go",
		"networking/outbound/messenger/friendnotification/packet.go",
		"networking/inbound/navigator/competition/random/packet.go",
		"networking/inbound/navigator/competition/room/packet.go",
		"networking/inbound/navigator/competition/search/packet.go",
		"networking/inbound/navigator/competition/submittable/packet.go",
	}
	for _, relative := range adjacent {
		identities = append(identities, readPacketIdentity(t, root, relative))
	}
	if len(identities) != 73 {
		t.Fatalf("complete packet surface has %d pairs, want 73", len(identities))
	}
	seen := make(map[packetIdentity]struct{}, len(identities))
	for _, identity := range identities {
		if _, duplicate := seen[identity]; duplicate {
			t.Fatalf("duplicate manifest identity %s:%d", identity.direction, identity.header)
		}
		seen[identity] = struct{}{}
	}
	for _, expected := range expectedCatalogPackets() {
		if _, found := seen[expected]; !found {
			t.Errorf("missing catalog identity %s:%d", expected.direction, expected.header)
		}
	}
}

// repositoryRoot resolves the checkout root from this test source file.
func repositoryRoot(t testing.TB) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve manifest source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}

// progressionPackets scans the capability-owned packet tree.
func progressionPackets(t testing.TB, root string) []packetIdentity {
	t.Helper()
	identities := make([]packetIdentity, 0, 63)
	for _, direction := range []string{"inbound", "outbound"} {
		base := filepath.Join(root, "networking", direction, "progression")
		err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() || entry.Name() != "packet.go" {
				return walkErr
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			identities = append(identities, readPacketIdentity(t, root, relative))
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Slice(identities, func(left int, right int) bool {
		if identities[left].direction == identities[right].direction {
			return identities[left].header < identities[right].header
		}
		return identities[left].direction < identities[right].direction
	})
	return identities
}

// readPacketIdentity reads one packet declaration from a relative source path.
func readPacketIdentity(t testing.TB, root string, relative string) packetIdentity {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, relative))
	if err != nil {
		t.Fatal(err)
	}
	match := manifestHeaderPattern.FindSubmatch(data)
	if len(match) != 2 {
		t.Fatalf("missing header declaration in %s", relative)
	}
	value, err := strconv.ParseUint(string(match[1]), 10, 16)
	if err != nil {
		t.Fatal(err)
	}
	direction := "inbound"
	if strings.Contains("/"+filepath.ToSlash(relative)+"/", "/outbound/") {
		direction = "outbound"
	}
	return packetIdentity{direction: direction, header: uint16(value)}
}

// expectedCatalogPackets returns the exact 24+33 catalog identities.
func expectedCatalogPackets() []packetIdentity {
	inbound := []uint16{196, 219, 359, 389, 793, 1190, 1296, 1334, 1364, 1371, 1782, 2077, 2127, 2397, 2399, 2486, 2595, 2750, 2912, 3077, 3133, 3144, 3333, 3604, 3720}
	outbound := []uint16{66, 133, 230, 305, 638, 740, 806, 949, 1066, 1122, 1177, 1203, 1689, 1745, 1878, 1968, 2107, 2265, 2501, 2589, 2665, 2772, 2927, 2998, 3027, 3370, 3406, 3506, 3625, 3841, 3926, 3954}
	values := make([]packetIdentity, 0, len(inbound)+len(outbound))
	for _, header := range inbound {
		values = append(values, packetIdentity{direction: "inbound", header: header})
	}
	for _, header := range outbound {
		values = append(values, packetIdentity{direction: "outbound", header: header})
	}
	return values
}
