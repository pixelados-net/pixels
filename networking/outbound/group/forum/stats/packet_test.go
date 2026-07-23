package stats

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
	"time"
)

// TestEncodeWritesExtendedForumData verifies settings and errors.
func TestEncodeWritesExtendedForumData(t *testing.T) {
	packet, err := Encode(grouprecord.ForumSummary{Group: grouprecord.Group{ID: 7}}, [5]string{}, true, false, time.Now())
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
