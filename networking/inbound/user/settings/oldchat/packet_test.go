package oldchat

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies old-chat decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Bool(true))
	payload, err := Decode(packet)
	if err != nil || !payload.OldChat {
		t.Fatalf("decode old chat: %#v, %v", payload, err)
	}
}
