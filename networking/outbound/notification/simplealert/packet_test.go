package simplealert

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the titled alert shape.
func TestEncode(t *testing.T) {
	packet, err := Encode("body", "title")
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].String != "body" || values[1].String != "title" {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
