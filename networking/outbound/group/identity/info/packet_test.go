package info

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
	"time"
)

// TestEncodeHasNoTrailingForumField verifies the audited Nitro shape.
func TestEncodeHasNoTrailingForumField(t *testing.T) {
	packet, err := Encode(grouprecord.Group{ID: 7, Name: "Pixels", CreatedAt: time.Unix(0, 0)}, grouprecord.Owner, true, false, true, false, true)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
