package contextmenu

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode writes Nitro's negative room bot id.
func TestEncode(t *testing.T) {
	packet, err := Encode(7)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil || decodeErr != nil || values[0].Int32 != -7 {
		t.Fatalf("values=%#v encode=%v decode=%v", values, err, decodeErr)
	}
}
