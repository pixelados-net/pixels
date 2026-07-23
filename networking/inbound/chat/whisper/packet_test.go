package whisper

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies UNIT_CHAT_WHISPER decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.String("alice hello"), codec.Int32(2))
	payload, err := Decode(packet)
	if err != nil || payload.RecipientAndMessage != "alice hello" || payload.StyleID != 2 {
		t.Fatalf("unexpected payload=%#v err=%v", payload, err)
	}
	packet.Header++
	if _, err = Decode(packet); err == nil {
		t.Fatal("expected header rejection")
	}
}
