package issueclose

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the native close-notification shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(2, "cerrado")
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].Int32 != 2 || values[1].String != "cerrado" {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
