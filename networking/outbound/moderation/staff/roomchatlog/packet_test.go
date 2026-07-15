package roomchatlog

import (
	"github.com/niflaot/pixels/networking/outbound/moderation/chatrecord"
	"testing"
)

// TestEncodeUsesHeader verifies room log projection.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode(chatrecord.Record{})
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
