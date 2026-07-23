package stats

import (
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	"testing"
)

// TestEncode verifies MARKETPLACE_ITEM_STATS encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(marketcore.Stats{}, 1, 22)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
