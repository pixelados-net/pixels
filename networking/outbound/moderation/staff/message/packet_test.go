package message

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeUsesHeader verifies the protocol identifier.
func TestEncodeUsesModeratorMessageShape(t *testing.T) {
	packet, err := Encode("Please stop", "")
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.StringField, codec.StringField}, packet.Payload)
	if err != nil || len(rest) != 0 || values[0].String != "Please stop" || values[1].String != "" {
		t.Fatalf("values=%+v rest=%d err=%v", values, len(rest), err)
	}
}
