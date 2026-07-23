package actionresult

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeUsesHeader verifies the protocol identifier.
func TestEncodeUsesTargetUserID(t *testing.T) {
	packet, err := Encode(37, true)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.BooleanField}, packet.Payload)
	if err != nil || len(rest) != 0 || values[0].Int32 != 37 || !values[1].Boolean {
		t.Fatalf("values=%+v rest=%d err=%v", values, len(rest), err)
	}
}
