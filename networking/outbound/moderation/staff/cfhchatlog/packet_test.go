package cfhchatlog

import (
	"github.com/niflaot/pixels/networking/outbound/moderation/chatrecord"
	"testing"
)

// TestEncodeUsesHeader verifies evidence projection.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode(1, 2, 3, chatrecord.Record{})
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
