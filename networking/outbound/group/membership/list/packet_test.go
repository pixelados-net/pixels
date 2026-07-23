package list

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
)

// TestEncodeWritesMembershipEntries verifies group list data.
func TestEncodeWritesMembershipEntries(t *testing.T) {
	packet, err := Encode([]grouprecord.PlayerGroup{{Group: grouprecord.Group{ID: 7, Name: "Pixels"}, Favorite: true}})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
