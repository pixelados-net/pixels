package threads

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
	"time"
)

// TestEncodeWritesThreadPage verifies list fields.
func TestEncodeWritesThreadPage(t *testing.T) {
	packet, err := Encode(7, 0, []grouprecord.Thread{{ID: 8}}, time.Now())
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
