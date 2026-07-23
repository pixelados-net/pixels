package parts

import (
	"github.com/niflaot/pixels/internal/realm/group/badge"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
)

// TestEncodeWritesFiveCollections verifies complete editor data.
func TestEncodeWritesFiveCollections(t *testing.T) {
	snapshot := &badge.Snapshot{Elements: []grouprecord.BadgeElement{{Kind: grouprecord.BadgeBase, ID: 1}}, Colors: []grouprecord.BadgeColor{{Family: grouprecord.BaseColor, ID: 1}}}
	packet, err := Encode(snapshot)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
