package modify

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsMutation verifies filter mutation decoding.
func TestDecodeReadsMutation(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(9), codec.Bool(true), codec.String("spam"))
	payload, err := Decode(packet)
	if err != nil || !payload.Add || payload.Word != "spam" {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}
