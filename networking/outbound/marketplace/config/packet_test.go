package config

import (
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	"testing"
)

// TestEncode verifies MARKETPLACE_CONFIG encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(marketcore.Options{})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
