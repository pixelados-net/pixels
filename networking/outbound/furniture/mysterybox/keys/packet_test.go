package keys

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies both tracker color identifiers.
func TestEncode(t *testing.T) {
	packet, err := Encode("blue", "gold")
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].String != "blue" || values[1].String != "gold" {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
