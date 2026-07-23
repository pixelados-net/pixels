package list

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
	"time"
)

// TestEncodeWritesForumPage verifies page metadata.
func TestEncodeWritesForumPage(t *testing.T) {
	packet, err := Encode(0, 1, 0, []grouprecord.ForumSummary{{Group: grouprecord.Group{ID: 7}}}, time.Now())
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
