package request

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsUsername verifies REQUEST_FRIEND decoding.
func TestDecodeReadsUsername(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.StringField}, codec.String("alice"))
	payload, err := Decode(packet)
	if err != nil || payload.Username != "alice" {
		t.Fatalf("unexpected payload=%#v err=%v", payload, err)
	}
}
