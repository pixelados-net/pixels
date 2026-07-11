package style

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies USER_SETTINGS_CHAT_STYLE decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(9))
	styleID, err := Decode(packet)
	if err != nil || styleID != 9 {
		t.Fatalf("unexpected style=%d err=%v", styleID, err)
	}
	packet.Header++
	if _, err = Decode(packet); err == nil {
		t.Fatal("expected header rejection")
	}
}
