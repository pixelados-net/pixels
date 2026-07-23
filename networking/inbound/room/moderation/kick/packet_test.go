package kick

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsPlayerID verifies kick decoding.
func TestDecodeReadsPlayerID(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.PlayerID != 7 {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
}
