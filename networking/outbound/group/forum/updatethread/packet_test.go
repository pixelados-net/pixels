package updatethread

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
	"time"
)

// TestEncodeWritesForumUpdate verifies the protocol header.
func TestEncodeWritesForumUpdate(t *testing.T) {
	packet, err := Encode(7, grouprecord.Thread{ID: 8}, time.Now())
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
