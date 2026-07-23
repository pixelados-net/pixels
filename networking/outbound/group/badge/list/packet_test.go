package list

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
)

// TestEncodeWritesBadgePairs verifies relevant badge projection.
func TestEncodeWritesBadgePairs(t *testing.T) {
	packet, err := Encode([]grouprecord.PlayerGroup{{Group: grouprecord.Group{ID: 7, BadgeCode: "b001010"}}})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
