package shout

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies UNIT_CHAT_SHOUT decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.String("hello"), codec.Int32(4))
	payload, err := Decode(packet)
	if err != nil || payload.Message != "hello" || payload.StyleID != 4 {
		t.Fatalf("unexpected payload=%#v err=%v", payload, err)
	}
	packet.Header++
	if _, err = Decode(packet); err == nil {
		t.Fatal("expected header rejection")
	}
}
