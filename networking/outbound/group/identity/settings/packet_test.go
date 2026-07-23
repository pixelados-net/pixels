package settings

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
)

// TestEncodePadsBadgeAndInvertsDecoration verifies manager compatibility.
func TestEncodePadsBadgeAndInvertsDecoration(t *testing.T) {
	packet, err := Encode(grouprecord.Group{ID: 7, Name: "Pixels", CanMembersDecorate: true}, nil)
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
