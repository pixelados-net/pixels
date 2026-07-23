package initmsg

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeUsesNativeLimits verifies MESSENGER_INIT fields.
func TestEncodeUsesNativeLimits(t *testing.T) {
	packet, err := Encode(200, 200, 500)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || packet.Header != Header || values[2].Int32 != 500 || values[3].Int32 != 0 {
		t.Fatalf("unexpected packet=%#v values=%#v err=%v decode=%v", packet, values, err, decodeErr)
	}
}
