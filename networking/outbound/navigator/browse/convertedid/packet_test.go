package convertedid

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies converted identifier wire order.
func TestEncode(t *testing.T) {
	packet, err := Encode("room:130", 130)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	values, remaining, err := codec.DecodePayload(nil, Definition, packet.Payload)
	if err != nil || len(remaining) != 0 || values[0].String != "room:130" || values[1].Int32 != 130 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
