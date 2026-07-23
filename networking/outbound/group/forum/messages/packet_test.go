package messages

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
	"time"
)

// TestEncodeWritesPostPage verifies list fields.
func TestEncodeWritesPostPage(t *testing.T) {
	packet, err := Encode(7, 8, 0, []grouprecord.Post{{ID: 9}}, time.Now())
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
