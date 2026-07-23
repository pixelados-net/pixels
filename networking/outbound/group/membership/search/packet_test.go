package search

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
	"time"
)

// TestEncodeWritesMemberPage verifies result and pagination fields.
func TestEncodeWritesMemberPage(t *testing.T) {
	page := grouprecord.MemberPage{Group: grouprecord.Group{ID: 7, Name: "Pixels"}, Members: []grouprecord.Membership{{PlayerID: 8, Username: "alice", JoinedAt: time.Unix(0, 0), Role: grouprecord.Admin}}, PageSize: 14}
	packet, err := Encode(page, true)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
